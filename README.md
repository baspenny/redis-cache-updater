# Ebay Data Streaming - Cache updater

A simple HTTP webserver that updates the cache of mapping Ebay Seller and Google Merchant Center ids.


## Local development
You can run the application locally with docker-compose or without docker-compose. If you run it without docker-compose 
you will need to have a redis instance running.

### Docker-compose
Copy the `.env.example` to `.env` and fill in the required variables

```text
REDIS_HOST=<redis_host>
REDIS_PASSWORD=<redis_password>
REDIS_PORT=<redis_port>
```

### Local
Set the environment variables that are stated in the  `.env.example` file and load them in your terminal.
Also make sure that you have a redis instance running. 
After that you can start the applicaton with:

```shell
go run .
```
The http server will spin up and listen on port `8080`. You can now make calls to the server.

### API endpoints
#### Refresh Cache:
```
POST /refresh
```

```json
# Body
{
	"project_id": "<Project_id>",
	"dataset_id": "<Dataset_id>",
	"table_id": "<Table_id>",
	"market": "the market UK|DE|FR|IT"
}
```
Returns `200` if the cache was updated successfully.

## Deployment
The application is deployed on Google Cloud Run. The deployment is done with the `Makefile` script.
You can find the commands in the Makefile and they are self-explanatory.