#!/bin/bash
#
# Usage: $SELF start|stop|restart
#

set -e

case "$1" in
  'start')

        ;;

  'stop')

        ;;

  'restart')
    	$0 stop
        $0 start
        ;;

  *)
        #echo "Usage: $SELF start|stop|restart"
        #can not exit
        #exit 1
        ;;
esac

exec "$@"