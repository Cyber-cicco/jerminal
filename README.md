# Jerminal : Jenkins in the terminal

⚠️  This is a work in progress. 

## What is Jerminal ?

Jerminal is a pipeline framework that allow you to describe a pipeline of events you want
to execute with helper functions. It also provides a set of detailed event logs, and
a system of agents and schedules with fine grained customization.

But more than this, it also provides a set of out of the box functionnalities, like
a unix socket server, an integration with github hooks, and a production of reports in JSON files,
and soon to be MongoDB and SQLITE.

You can use it for all sorts of things: 
 * CI/CD
 * Deterministic simulation testing
 * Integration testing
 * Load balancing


## Why Jerminal ?

After an intense 2 days of using Jenkins and wanting to `rm -rf / --no-preserve-root` myself IRL,
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





