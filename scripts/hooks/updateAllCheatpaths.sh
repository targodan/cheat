#!/bin/bash

# This is used to only perform updates once a day.
# Adjust this if you wish.
# Note that on many distros /tmp is a tmpfs, meaning the file is lost if you reboot.
dailyFile="/tmp/cheat.daily"
today=$(date +%d.%m.%y)

if [ -f "$dailyFile" ]; then
    if [[ "$today" == $(cat "$dailyFile") ]]; then
        # Already executed today
        exit
    fi
fi

pwd >&2
echo "COUTN: $CHEAT_CONF_CHEATPATHS_COUNT" >&2

for i in $(seq 0 $((CHEAT_CONF_CHEATPATHS_COUNT-1))); do
    varname="CHEAT_CONF_CHEATPATHS_${i}_PATH"
    echo ${!varname} >&2
done

exit 1

echo $today > "$dailyFile"