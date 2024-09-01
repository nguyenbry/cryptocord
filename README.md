Simple web app that stores Discord webhook urls and periodically will send the current price of Bitcoin to them. Market data is provided by CoinMarketCap and data is stored in Postgres.

To get started:

- Install Go packages
- Spin up a Docker container with: `docker run --name my-container-name -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=mydbname -p 5432:5432 -d postgres`
- Create a `.env` file at the root of the project, at the same level of `go.mod`:

```ts
DB_URI = "postgres://postgres:mysecretpassword@localhost:5432/mydbname";
CMC_KEY = "cf9c22...";
```

- Use `goose` to migrate the database. Run from within the `sql/migrations` directory: `goose postgres postgres://postgres:mysecretpassword@localhost:5432/mydbname up`
- Start the app from the root directory: `env ENV=DEV go run .`
- Register your first receiving webhook by making a POST request to `http://localhost:4000/api/job` with payload:

```json
{
  "url": "mydiscordwebhook"
}
```

You should receive a message with the price of Bitcoin at this webhook (and any others in the db) on an interval. Edit the interval duration in `main.go` as needed.
