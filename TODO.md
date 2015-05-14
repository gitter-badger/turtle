## Daemon

* Turtle docker container ist still not terminating nice.

* command option
* Implement automatic backup with encrypted compressed export.
* Add exclude automatic backup option

* Check online service (network port pinging, https status requests...)
* Sort the backup list before sending it to the client.
* Implement logging features.
* add possibilities to limit resources for each app.
* log cpu, storage, memory usage of each app.
* Create a temporary testing clone of an app during an update.
* Implement access groups
* Validate the Turtlefile for invalid env.containes and port.container values.

### Optional
* Check if the host fingerprint is missing also on the update command?



## Client

* Implement an user-friendly export and import option.
* Implement awesome terminal info screen (https://github.com/gizak/termui)
  Currently commented in cmd_watch.go