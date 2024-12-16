# Labours-go

Labours-go is a project aimed at replacing the Python implementation of [labours](https://github.com/src-d/hercules/tree/master/python/labours) with a Go-based implementation for better performance, maintainability, and scalability. This repository contains the Go codebase and all the necessary tools to analyze and visualize contributions in Git repositories.

## Features

* High Performance: Leverages Go’s concurrency model for faster analysis of Git repositories.
* Ease of Use: Simple command-line interface to perform various operations (compatible with the original labours).
* Compatibility: Fully compatible with the data analysis and metrics provided by the original Python version.
* Extensible: Modular and extensible design to allow further customization and enhancement.

> **Note:** This project is still under development and not yet not working as it should. It's rather a proof-of-concept.

## Installation

### Prerequisites

* Go version 1.18 or higher.
* Git installed on your machine.

### Steps

1. Clone the repository:

```bash
git clone https://github.com/MeKo-Christian/labours-go.git
cd labours-go
```

1. Build the project:

```bash
go build -o labours
```

1. Verify installation:

```bash
./labours-go
```

