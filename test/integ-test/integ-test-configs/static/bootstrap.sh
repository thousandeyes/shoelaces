#!/bin/bash
export DEBIAN_FRONTEND=noninteractive

echo Example boostrap configuration
apt-get install hello

echo '#!/bin/sh
exit 0' > /etc/rc.local

# Don't want to run this accidentally.
chmod 0 /usr/local/sbin/bootstrap
