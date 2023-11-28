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

func insertInRedis(ctx context.Context, p *redis.Pool, d *EbayGMCLookup, market string) error {
	conn := p.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	_, err := conn.Do("SET", market+":"+d.Brand, d.GmcAccountId)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func GetStats(ctx context.Context) (string, error) {
	pool, err := Pool()
	if err != nil {
		return "", err
	}
	conn := pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	stats, err := redis.String(conn.Do("MEMORY STATS"))
	fmt.Println(stats)
	if err != nil {
		return "", err
	}
	return stats, nil
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

	counter := 0

	for {
		var ebayGMCLookup EbayGMCLookup
		err := it.Next(&ebayGMCLookup)
		if errors.Is(err, iterator.Done) {
			fmt.Println("All results processed...")
			wg.Wait()
			break
		}
		if err != nil {
			//panic("Cannot read data" + err.Error())
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = insertInRedis(ctx, pool, &ebayGMCLookup, market)
			if err != nil {
				log.Warningf("Could not insert record: %v", err.Error())
			}

		}()
		if counter == 250 {
			fmt.Print(".")
			wg.Wait()
			// reset the counter
			counter = 0
		}
		counter++
	}

	// Stop the clock and clean up!
	elapsed := time.Since(start)
	fmt.Println("Took ", elapsed)
	err = pool.Close()
	if err != nil {
		return err
	}
	return nil
}
