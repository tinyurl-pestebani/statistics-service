Statistics Service
==================

This gRPC service contains the implementation for `StatisticsService` defined in the https://github.com/tinyurl-pestebani/proto/blob/main/v1/statistics.proto file.


Features
--------

*   **gRPC Interface**: Exposes a gRPC interface for statistics retrieval.

*   **Database Integration**: Connects to a database to fetch statistics data.

*   **Observability**: Integrated with OpenTelemetry for tracing and logging.

*   **Configuration**: The service can be configured using environment variables.

*   **Containerized**: Comes with a Dockerfile for easy containerization and deployment.


Getting Started
---------------

### Prerequisites

*   Go (version 1.24.5 or higher)

*   Docker and Docker Compose


### Building and Running

You can build and run the service using the provided Dockerfile.

To build the Docker image, run the following command from the root directory:

`docker build -t statistics-service:0.0.1 .`


Environment Variables
---------------------

The following environment variables can be used to configure the service:

*   **STATISTICS_SERVICE_PORT**: The port on which the gRPC server listens. Default value is **`8082`**


For setting OpenTelemetry, and the database, please refer to https://github.com/tinyurl-pestebani/go-otel-setup and https://github.com/tinyurl-pestebani/statistics-database packages.

API
---

The service provides the following gRPC methods:

*   **GetStatistics**: Retrieves statistics for a given tag within a specified time interval.

*   **Ping**: A simple health check endpoint that returns "pong".
