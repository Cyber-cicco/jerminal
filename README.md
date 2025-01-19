# Jerminal : Jenkins in the terminal

⚠️  This is a work in progress. 

## What is Jerminal ?

It's Jenkins in the form of a go library with whom you can do everything jenkins does
by just writing go code, compiling it and running. It's super fast, and allows you
to do administration of your CI/CD without exposing an endpoint of your server with
privileged access to a lot of critical resources, by just letting you write your pipeline
locally and deploy it through SSH.

## Why Jerminal ?

After an intense 2 days of using Jenkins and wanting to `rm -rf / --no-preserve-root` IRL,
I came to the conclusion that a web interface was probably not the best way to do what Jenkins
is doing.

The idea behind Jerminal is simple : there is not a single thing in jenkins that wouldn't be easier
if it was done through code and a terminal. Installing a tool, a plugin, scripting the pipeline, all
of this would be so much more easier in an IDE with config files and a scripting language that enables
type safety. That what jerminal does.

## How to use Jerminal ?

Set up a linux environnement with the required dependencies to build your project.
Write a go script that describes your pipeline and launches the Jerminal server.
Tada, you have your CI / CD without having to bother with Jenkins.

Check main.go to have a overview of what i would want to acheive with the API





