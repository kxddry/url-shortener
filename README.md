# url-shortener
REST API go url shortener

# url-shortener
REST API go url shortener

# Overview
This project is a URL shortener service implemented as a REST API using the Go programming language. It allows users to shorten long URLs and retrieve the original URLs using the shortened versions.

# Features
- RESTful API for URL shortening and retrieval
- Persistent storage for shortened URLs
- Configurable storage backends (PostgreSQL, Redis by default)
- Scalable and efficient design
- Integrated with [sso-auth](https://github.com/kxddry/sso-auth)
- Fully functioning authorization and authentication system.

# Getting Started

## Usage
1. Clone the repository:
   ```bash
   git clone
   ```
2. Set up the configuration file:
    - Create a `config.yaml` file wherever needed.
    - Define the necessary configuration parameters (refer to the example configuration file).
    - Make sure to set the CONFIG_PATH environment variable to point to the config.yaml file.

3. Run the application:
   ```bash
   task run
   ```
4. Access the API:
    - Shorten a URL:
      ```
      POST /url
      {
         "url": "https://example.com/long-url",
         "alias": "short-url" # can be omitted
      }
      ```
   - Retrieve the original URL:
     ```
     GET /{shortened_url}
     ```
   - Login:
    ```
     POST /login
     {
         "placeholder": "(username or email)",
         "password": "(password)"
      }
     ```
   - Register:
   ```
   POST /register
     {
         "username": "username",
         "email": "email@g.co",
         "password": "password"
      }
   ```
   - Delete:
   ```
   DELETE /{alias} (with JWT bearer token in headers, no JSON required)
   ```
The alias will only be deleted if it was created by the same user or if an admin is trying to delete it.
## todo:
- [ ] Add more tests
- [X] implement Redis
- [ ] debug Redis
- [X] add authorization
