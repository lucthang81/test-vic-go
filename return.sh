if pgrep "vic_go" > /dev/null
then
    :
else
    cd "/home/tungdt/go/src/github.com/vic/vic_go" && "./vic_go"
fi

# cron command:
#    */1 * * * * sh /home/tungdt/go/src/github.com/vic/vic_go/return.sh 1>vicGoOut.txt 2>vicGoOut.txt
