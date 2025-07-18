# Project Setup

This document provides detailed instructions on setting up the Event Shark project.

## Prerequisites

Before setting up the project, ensure you have the following prerequisites installed:

- Docker
- Docker Compose
- Git

## Dependencies

The project relies on the following dependencies:

- Kafka
- Schema Registry
- Redpanda

## Step-by-Step Setup Instructions

1. **Clone the repository:**
   ```sh
   git clone https://github.com/dipjyotimetia/EventShark.git
   cd EventShark
   ```

2. **Build the Docker images:**
   ```sh
   docker-compose build
   ```

3. **Start the services:**
   ```sh
   docker-compose up -d
   ```

4. **Verify the setup:**
   - Check if the services are running:
     ```sh
     docker-compose ps
     ```
   - Ensure the application is healthy by accessing the health endpoint:
     ```sh
     curl http://localhost:8083/health
     ```

5. **Access the Kafka console:**
   - Open your browser and navigate to `http://localhost:8086` to access the Kafka console.

6. **Create topics and schemas:**
   - The topics and schemas are automatically created during the setup process. You can verify them using the Kafka console.

7. **Run the tests:**
   - To run the tests, use the following command:
     ```sh
     make test
     ```

8. **Build the project:**
   - To build the project, use the following command:
     ```sh
     make build
     ```

9. **Clean up:**
   - To clean up the Docker containers and images, use the following command:
     ```sh
     make clean
     ```

## AVSC to JSON Utility

If you need to convert Avro schemas to JSON for use with the Schema Registry, use the utility in `script/avsc2json`. See [script/avsc2json/readme.md](../script/avsc2json/readme.md) for details.

## Troubleshooting

- If Docker containers fail to start, check logs with `docker-compose logs`.
- If topics or schemas are missing, ensure the `init-topic` service completes successfully.
- For port conflicts, ensure ports 8083, 8086, 8081, 8082, 9092, and 29092 are free.
- For schema errors, verify your Avro schemas and use the AVSC to JSON utility.

By following these steps, you should have the Event Shark project set up and running on your local machine.
