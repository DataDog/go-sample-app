# go-sample-app
A sample application composed of several services, written in Go.

This application is a downloadable, interactive example that demonstrates the use of Datadog's Go auto-instrumentation ([DataDog/orchestrion](https://github.com/DataDog/orchestrion))

The application is a very basic implementation of a multi-user notes app that allows users to submit and read notes.

## Getting Started

To get started with this sample application, we can first launch the application and take a look at it.
It is a very simple example web application with a *very* basic UI, but this is sufficient to demonstrate Orchestrion's use.

To start with, make sure you have an [API Key](https://docs.datadoghq.com/account_management/api-app-keys/) ready so that traces can appear in the UI.
You should export a variable called `DD_API_KEY` in your shell environment. This will be picked up by the Datadog Agent and used to submit traces.

```shell
$ export DD_API_KEY=[your api key]
```

Next we can bring up the application. It starts out with no instrumentation present. We'll just click through the app to see how it works and what it does. (There's not much).
By default, the applications bind to `localhost:8080` and `localhost:8081`, so make sure those ports are not occupied by other services on your machine.

If you cannot free up those ports, you can modify the `docker-compose.yml` file in the root directory of this project, changing the port bindings in the `ports:` section.

ex: moving the `users` service from `8080` to `8083`:
```diff
  users:
    container_name: users
    build: ./services/users
    ports:
-     - "8080:8080"
+     - "8083:8080"
```
Note: The rest of this document assumes you're on the default ports.


Now we build and start the services:
```shell
$ docker-compose build
$ docker-compose up -d
```
If you go to [`http://localhost:8080`](http://localhost:8080) in your browser, you should see the application's main page. This is served by the `user` service in the `services/user` directory.

It is a simple directory of the imaginary users of our notes application.
![Application Root](doc/root.png)

Clicking on a user's ID will take you to their "notes" page, where you can read and add notes.
![User Notes Page](doc/user-notes.png)

Try adding a note or two.

These notes are read from and saved to the `notes` service in the `services/notes` directory.

If you want, take a little time to look at the applications in [services/users/main.go](services/users/main.go) and [services/notes/main.go](services/notes/main.go).

## Adding Instrumentation
So far we have just run and played with the application, but there is no instrumentation currently in it. This means no traces are created or sent to Datadog.

We want visibility into our application and what it's doing via Datadog APM. This will allow us to see traces of the requests that come into the system and analyze the flow and time taken to fulfil a request.

With `orchestrion`, this instrumentation is easy to add. 

First, make sure you have `orchestrion` installed:
```sh
$ go install github.com/datadog/orchestrion@latest
```

Next, we will simply run orchestrion over our code base. This will recursively go through the directories, finding go files and adding instrumentation code to them.
```sh
$ orchestrion -w .
Scanning Package /.../go-sample-app
overwriting /.../go-sample-app/services/notes/main.go:
overwriting /.../go-sample-app/services/users/main.go:
overwriting /.../go-sample-app/tools/header_check.go:
```

If you are curious about the changes orchestrion made, take a look at the git diff:
```sh
$ git diff
...
```

You should see new function calls and wrappers placed around relevant code such as http handler functions and sql clients. You'll also see comments added, such as `//dd:startwrap`. These comments are used to identify instrumented code, so that orchestrion can be run multiple times without duplicating instrumentation, and also so orchestrion can remove instrumentation when instructed to with the `-rm` flag.

Now, shut down the services:
```
$ docker-compose down
```

We need to update our go.mod files before we can build the new code. The `Orchestrion` instrumentation requires code from the `github.com/DataDog/orchestrion` library.

```sh
$ cd services/users
services/users $ go mod tidy
go: finding module for package github.com/datadog/orchestrion/instrument
go: found github.com/datadog/orchestrion/instrument in github.com/datadog/orchestrion v0.1.0
services/users $ cd ../notes/
services/notes $ go mod tidy
go: finding module for package github.com/datadog/orchestrion/instrument
go: found github.com/datadog/orchestrion/instrument in github.com/datadog/orchestrion v0.1.0
services/notes $ cd ../../
$ 
```

Then rebuild the applications with our newly instrumented code
```sh 
$ docker-compose build
```

And start the services again:
```sh 
$ docker-compose up -d
```

Now, going to the application home page [http://localhost:8080](http://localhost:8080) will show us the same page, but looking at our Traces Page in the Datadog app [https://app.datadoghq.com/apm/traces](https://app.datadoghq.com/apm/traces) should show some new traces appear!

![First Trace](doc/trace1.png)

Click through the various spans and tabs. For instance, by clicking the `sqlite3.query` span, we can see the SQL query and other info.
![SQLite3 Trace Data](doc/sqlite-info.png)


## More Traces
Now that we have instrumentation, let's generate some more interesting traces. From the browser, click on one of the users' IDs to go to their notes page.

Navigating to that page should generate another trace, which looks like this:
![Notes Page](doc/notes-trace.png)

Here we can see that the `users` service received an http request. It then sent an outgoing http request to the `notes` service, requesting `/notes?userid=2` in this case (this will depend on which user's page you are viewing).
![Outgoing](doc/notes-trace2.png)

The `notes` service then executes a SQL query:
![Notes SQL Query](doc/notes-sql.png)


Try adding a new note. When you do that, you should see a new trace appear.
![New Note Trace](doc/notes-trace3.png)
This trace appears with type `POST` and status code `302` because the `users` service returns a redirect to the browser after the submission succeeds.

Once again, we can see in the trace that the `users` service talks to the `notes` service, this time resulting in an `INSERT` query.
![New Note Trace Details](doc/notes-trace4.png)

## Removing Instrumentation
Now that we've seen how to add instrumentation to the code, we might as well cover how to remove it. 

Removing instrumentation is as easy as adding it. If, for whatever reason, you want to remove the instrumentation added by `orchestrion`, run the following in the repository root:
```sh 
$ orchestrion -rm -w .
Scanning Package /.../go-sample-app
Removing Orchestrion instrumentation.
overwriting /.../go-sample-app/services/notes/main.go:
overwriting /.../go-sample-app/services/users/main.go:
overwriting /.../go-sample-app/tools/header_check.go:
```

Next, run `go mod tidy` again, as we did before, for both services. This will clean up the dependencies we had to add earlier.

A git diff should show that the instrumentation has been removed.
```sh
$ git diff
```
