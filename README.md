# Golang-httpflood ![](https://img.shields.io/badge/Version-2.0-brightgreen.svg) ![](https://img.shields.io/badge/license-MIT-blue.svg)
Using Golang(net/socket) to httpflood

**Warning: Please use command "ulimit -n 999999" before use this in linux**

**1 Threads =  1 connection, 100~300 connections can down a normal website in 10s(specially apache server LOL)**
 
**This is golang and threads are just goroutines so you set more higher threads like 1000-5000 is fine.**

**Why can it run over 1000 threads(goroutines)? [Read this](http://tleyden.github.io/blog/2014/10/30/goroutines-vs-threads/)**

## INFO

 - [x] HTTP Get Flood
 - [x] HTTP Post Flood
 - [x] Random url(http get flood)
 - [x] Self edit header(You can use "nil" to use default header)
 - [x] Improved threading control
 - [x] More powerful flood
 - [x] Auto get ip form domain(golang inbuilt function)
 - [x] More format for random url(http get flood)
 - [x] Fixed for 386 systems
 -----------------------------------------------------
 Default header setting:
 - [x] Random user-agents
 - [x] Random data(http post flood) 
 - [x] Random Accpetall
 - [x] Random Referer(only for http get flood)


## Download
***Please download the F\*cking golang at first.***

Then:

    git clone https://github.com/Leeon123/golang-httpflood.git

Header.txt format:

    Accept: text/html
    User-agent: Wget
    Referer: http://google.com

Or anything else of http header. If you don't have any idea of this please just use "nil" for using default random header.
## Usage

    cd golang-httpflood
    go build httpflood.go
    ./httpflood  <url> <threads> <get/post> <seconds> <header.txt/nil>

## Docker production setup

A production-ready Docker image is now included.

Build the image locally:

    docker build -t golang-httpflood .

Run the container:

    docker run -p 8080:8080 --name golang-httpflood golang-httpflood

Then open:

    http://localhost:8080

If port 8080 is in use, map to another host port:

    docker run -p 8081:8080 --name golang-httpflood golang-httpflood

## Docker Compose

A `docker-compose.yml` is included for convenience:

    docker-compose up --build

This will expose the UI on port 8080 by default.

The web UI stores run state and logs in SQLite. By default the local binary
uses `httpflood.sqlite` in the working directory. In Docker Compose, the DB is
stored at `/data/httpflood.sqlite` and mounted to `./data`.

Optional web UI environment variables:

    HTTPFLOOD_DB=/path/to/httpflood.sqlite
    HTTPFLOOD_ADDR=:8080
    HTTPFLOOD_ADMIN_USER=admin
    HTTPFLOOD_ADMIN_PASS=admin123

## Authentication and account management

- Web UI now requires login (`/login`).
- On first startup, if no admin exists in SQLite, the app creates exactly one
  admin account from `HTTPFLOOD_ADMIN_USER` and `HTTPFLOOD_ADMIN_PASS`.
- Only admin can open `/accounts` to create and manage other accounts.
- Managed accounts are `member` role only. Admin role is unique and not created
  from UI.
- Account permissions:
  - `can_start_run`: can create flood runs.
  - `can_view_monitor`: can view VPS metrics page.
- Realtime monitor page is available at `/monitor` and updates without full page
  reload.

## VPS deployment

On your VPS, you can deploy from the repo using Docker:

    git clone <repo-url> golang-httpflood
    cd golang-httpflood
    docker build -t golang-httpflood .
    docker run -d -p 8080:8080 --name golang-httpflood golang-httpflood

Then access the service at your VPS public IP on port 8080.

> I cannot connect to your VPS or use SSH credentials directly, but these are the commands you can run on the server.
