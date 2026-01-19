NAME:
   go-stress-test - A high-performance HTTP load testing tool

USAGE:
   go-stress-test [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url value, -u value                                  Target URL to stress test (e.g., http://localhost:8080/api)
   --method value, -m value                               HTTP method (GET, POST, PUT, DELETE, etc.) (default: "GET")
   --body value                                           Request body as a string
   --body-file value                                      Path to file containing request body (overrides --body)
   --header value, -H value [ --header value, -H value ]  Custom HTTP headers (can be used multiple times, e.g., -H "Content-Type: application/json")
   --rate value, -r value                                 Requests per second (RPS) (default: 0)
   --concurrency value, -c value                          Number of concurrent workers (goroutines) (default: 10)
   --duration value, -d value                             Duration of the test (e.g., 10s, 1m, 2h) (default: 0s)
   --total value, -n value                                Total number of requests to send (overrides --duration if > 0) (default: 0)
   --timeout value                                        Timeout for each request (default: 30s)
   --output value, -o value                               Output file for report (use 'stdout' for terminal)
   --quiet, -q                                            Suppress progress output (default: false)
   --disable-keepalive                                    Disable HTTP Keep-Alive connections (use short-lived connections) (default: false)
   --help, -h                                             show help