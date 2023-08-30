# Order Notification System Readme

## Introduction
Welcome to the Order Notification System project! This readme file provides an overview of the system's key functionalities and APIs.

## Functionality Overview
The Order Notification System is designed to facilitate order tracking and investigation. It includes the following APIs:

### Order Delay Notification API
This API allows users to report a delay in an order after the "time_delivery" period has elapsed. Users can provide the necessary parameters for the delay report, such as order ID and reason for the delay. The API associates the delay report with the corresponding order, enabling efficient tracking of delays.

### Dedicate Agent API
The API sends a notification to a free agent, providing the necessary details of the order they need to handle and assign the delay report to them.

### Receive Vendor Delay Reports API
The API returns a list of vendors, ordered by their overall delay reports, providing valuable insights into vendor performance.

## Getting Started
To start using the Order Notification System and its APIs, follow these steps:

1. Clone the project repository to your local machine.
2. Set up the required dependencies and ensure they are properly installed.


### Migrating the Database
To migrate the database using Docker, run the following command:

```shell
docker compose --profile=migration up migrate postgres
```

### Building Adminer
To build Adminer using Docker, run the following command:

```shell
docker-compose --profile=dbadmin up -d adminer
```

### Building PostgreSQL
To build PostgreSQL using Docker, run the following command:

```shell
docker-compose up -d postgresql
```

### Building RabbitMQ
To build RabbitMQ using Docker, run the following command:

```shell
docker-compose up -d rabbitmq
```