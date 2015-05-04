## Daemon

* keep the authorized keys file in the docker container. make it persistant
* Implement automatic backup with encrypted compressed export.
* Add exclude automatic backup option

* Add possibility to build docker image from source. 
* Check if the host fingerprint is missing also on the update command.
* Sort the backup list before sending it to the client.
* Implement logging features.
* Create a temporary testing clone of an app during an update.
* Implement access groups
* Validate the Turtlefile for invalid env.containes and port.container values.




## Client

* Implement awesome terminal info screen (https://github.com/gizak/termui)
  Currently commented in cmd_watch.go