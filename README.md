# In-Memory Key-Value Database

This in-memory key-value database allows performing operations on it using a REST API and supports various commands such as SET, GET, QPUSH, QPOP, and BQPOP.

## TLDR

1. Open a terminal and navigate to the root directory of the project.

2. Compile the code
    ```shell
    go build cmd
    ```
3. Run the executable
   ```shell
   ./cmd
   ```
4. Run the tests
   ```shell
   go run -v ./tests/
   ```

## Code Structure

The code is organized into multiple directories in following fashion:

- `cmd/`
  - `main.go`: Contains the main entry point of the application, including the HTTP server setup and route handling.
- `database/`
  - `database.go`: Defines the `Database` struct and its associated methods, including `NewDatabase`, `Set`, `Get`, `QPush`, `QPop` and `BQPop`. `startExpiryCleanup` function handles the expiry cleanup functionality.
- `handlers/`
  - `http_handler.go`: Implements the HTTP request handlers for the database commands.
- `commandparser/`
  - `commandparser.go`: This function implements command parsing in a user-friendly manner, while also checking for continuous spaces and disregarding them. It also includes error handling to address cases of malformed commands or incorrect numbers of arguments being passed.



## Database Functionality

### SET Command

The `SET` command writes a value to the database based on the specified key and parameters. It supports optional fields such as expiry time and condition. Here's the pattern for the `SET` command:

`SET <key> <value> <expiry time>? <condition>?`


- `<key>`: The key under which the value will be stored.
- `<value>`: The value to be stored.
- `<expiry time>` (optional): Specifies the expiry time of the key in seconds.
- `<condition>` (optional): Specifies the decision to take if the key already exists. Accepts either `NX` or `XX`.

### GET Command

The `GET` command retrieves the value stored using the specified key. Here's the pattern for the `GET` command:

`GET <key>`


- `<key>`: The key for which to retrieve the value.

### QPUSH Command

The `QPUSH` command creates a queue if it doesn't already exist and appends values to it. Here's the pattern for the `QPUSH` command:

`QPUSH <key> <value...>`


- `<key>`: The name of the queue to write to.
- `<value...>`: Variadic input that receives multiple values separated by space.

### QPOP Command

The `QPOP` command returns the last inserted value from the queue. Here's the pattern for the `QPOP` command:

`QPOP <key>`


- `<key>`: The name of the queue.

### BQPOP Command

The `BQPOP` command is a blocking queue read operation that wait for a timeout period, if the element is available in the queue before the timeout, the element is returned otherwise a empty value after the timeout period.
 `BQPOP` command:

`BQPOP <key> <timeout>`


- `<key>`: The name of the queue to read from.
- `<timeout>`: The duration in seconds to wait until a value is available from the queue.

## Expiry Cleanup

The expiration functionality automatically removes expired keys from the database. Here's how it works:

- When a new instance of the database is created using `NewDatabase()`, the `startExpiryCleanup` method is called to start the expiry cleanup process.
- The expiry cleanup process runs as a goroutine and periodically checks for expired keys using a `Ticker` with a time interval of 1 second.
- When an expired key is detected, it is removed from the database.


