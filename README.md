# Company Earnings
[![GitHub License](https://img.shields.io/badge/license-GPL--2.0-blue.svg)](https://github.com/api-moose/company-earnings/blob/main/LICENSE)

⭐️ If you find this project useful, please star it on GitHub!

## About
The Financial Data Platform is a robust framework designed to process and serve financial data through a well-structured API. This platform includes features for managing users, stocks, and company information, making it ideal for building financial applications, trading systems, or market analysis tools.

## Table of Contents
- [Pre-requisites](#pre-requisites)
- [Installation](#installation)
- [Build and Run](#build-and-run)
- [Development Mode](#development-mode)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## Pre-requisites
Before proceeding, ensure that your system meets the following requirements:
- Go (version 1.22.4 or later)
- Docker
- Docker Compose

## Installation
1. Clone the repository:
   ```
   git clone https://github.com/yourusername/financial-data-platform.git
   cd financial-data-platform
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Build and Run
To build and run the Financial Data Platform:

1. Build the project:
   ```
   go build -o financial-data-platform ./cmd/api
   ```

2. Run the application:
   ```
   ./financial-data-platform
   ```

Alternatively, you can use the provided Makefile:
```
make build
make run
```

## Development Mode
To run the application in development mode with live reloading:

1. Install [Air](https://github.com/cosmtrek/air):
   ```
   go install github.com/cosmtrek/air@latest
   ```

2. Run the application using Air:
   ```
   air
   ```

## Testing
To run the tests for the Financial Data Platform:

```
go test ./...
```

For more specific test runs or to generate coverage reports, refer to the testing section in the project documentation.

## Project Structure
```
financial-data-platform/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── models/
│   │   ├── user.go
│   │   ├── stock.go
│   │   └── company.go
│   ├── services/
│   └── utils/
├── pkg/
├── tests/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

## Contributing
We welcome contributions to the Financial Data Platform! Please read our [CONTRIBUTING.md](CONTRIBUTING.md) file for details on our code of conduct and the process for submitting pull requests.

## License
This project is licensed under the GNU General Public License v2.0 - see the [LICENSE](LICENSE) file for details.

© 2024 YourCompanyName