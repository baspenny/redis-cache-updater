package cache

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"fmt"
	"github.com/apsystole/log"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/api/iterator"
	"sync"
	"time"
)

type EbayGMCLookup struct {
	Brand        string
	GmcAccountId string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func insertInRedis(p *redis.Pool, d *EbayGMCLookup, market string) error {
	conn := p.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("Failed to close Redis connection: %v", err)
		}
	}(conn)

	_, err := conn.Do("SET", market+":"+d.Brand, d.GmcAccountId)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	// Set the cache to expire after 8 days
	expireTime := 8 * 86400 * time.Second
	_, err = conn.Do("EXPIRE", market+":"+d.Brand, expireTime.Seconds())
	if err != nil {
		return err
	}

	return nil
}

func RefreshRedisCache(ctx context.Context, market string) error {
	mapper := map[string]string{
		"IT": "EBAY_IT",
		"FR": "EBAY_FR",
		"DE": "EBAY_DE",
		"UK": "EBAY",
	}

	start := time.Now()
	pool, err := Pool()
	if err != nil {
		return err
	}
	// Removed this because it was causing the closing of the pool before the goroutines could finish
	//defer func(pool *redis.Pool) {
	//	err := pool.Close()
	//	if err != nil {
	//		log.Warningf("Could not close Redis pool: %v", err.Error())
	//	}
	//}(pool)

	client, err := bigquery.NewClient(ctx, "nmpi-feeds")
	if err != nil {
		return err
	}
	queryString := fmt.Sprintf(`
		SELECT gmc_account_id as GmcAccountId, brand FROM nmpi-feeds.FEED_TEMP_TABLES.%s
		GROUP BY brand, gmc_account_id ORDER BY gmc_account_id`, mapper[market])

	q := client.Query(queryString)

	it, err := q.Read(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	total_count := 0
	counter := 0

	for {
		var ebayGMCLookup EbayGMCLookup
		err := it.Next(&ebayGMCLookup)
		if errors.Is(err, iterator.Done) {
			log.Infof("Done processing %d rows", it.TotalRows)
			wg.Wait()
			break
		}
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = insertInRedis(pool, &ebayGMCLookup, market)
			if err != nil {
				log.Warningf("Could not insert record: %v", err.Error())
			}

		}()

		if counter == 100 {
			wg.Wait()
			if total_count%10000 == 0 && total_count != 0 {
				log.Infof("%d records from inserted for market %s", total_count, market)
			}
			// reset the counter
			total_count += counter
			counter = 0
		}
		counter++

	}
	wg.Wait()
	// Stop the clock and clean up!
	elapsed := time.Since(start)
	log.Infof("Updating cache done for market %s. It took %v", market, elapsed)

	return nil
}
