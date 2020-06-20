# Jim Gilbert's Wheels and Deals web application

This application is the main site for Jim Gilbert's Wheels and deals.
It is written as an extension to [goBlender](https://github.com/tsawler/goblender).

## Development Setup

Clone [goBlender](https://github.com/tsawler/goblender), and then clone
this repository into `./ui/client/clienthandlers`

Add PBS and other authentications to the `.env` file:

~~~.env
PBSUSER=username
PBSPASS=somepassword

CARGURUHOST=ftp.cargurus.com
CARGURUSUSER=
CARGURUSPASS=

KIJIJIHOST=ftp.cargigi.com
KIJIJIUSER=
KIJIJIPASS=

KIJIJIPSHOST=cargigi.com
KIJIJIPSUSER=
KIJIJIPSPASS=
~~~