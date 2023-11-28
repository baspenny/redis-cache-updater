# Cache updater

This is a simple script that will update the cache of Google Merchant Center.


## Usage
### Docker-compose
Copy the `.env.example` to `.env` and fill in the required variables
```text
REDIS_HOST=<redis_host>
REDIS_PASSWORD=<redis_password>
REDIS_PORT=<redis_port>
```
### Local
set the environment variables that are stated in the  `.env.ecxmple` file and load them in your terminal.  
After that you can start the applicaton with:

```shell
go run .
```
The http server will spin up and listen on port `8080`. You can now make calls to the server.

### API calls
The API is very simple and only has two endpoints.

```json
{
	"project_id": "nmpi-feeds",
	"dataset_id": "FEED_TEMP_TABLES",
	"table_id": "EBAY_UK_11232",
	"market": "the market UK|DE|FR|IT"
}
```


