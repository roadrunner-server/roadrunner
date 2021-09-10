# Building a Server

RoadRunner use Endure to manage dependencies, this allows you to tweak and extend application functionality for each separate project.

#### Install Golang

To build an application server you need [Golang 1.16+](https://golang.org/dl/) to be installed.

#### Step-by-step walkthrough  

1. Fork or clone [roadrunner-binary](https://github.com/spiral/roadrunner-binary/) repository.
2. Feel free to modify plugins list or add your custom plugins to the [Plugins](https://github.com/spiral/roadrunner-binary/blob/master/internal/container/plugins.go).

You can now start your server without building `go run cmd/rr/main.go serve`.

> See how to create [http middleware](/http/middleware.md) in order to intercept HTTP flow.
