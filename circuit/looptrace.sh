RED='\033[0;31m'
LGREEN='\033[1;32m'
LBLUE='\033[1;34m'
DGRAY='\033[1;30m'
NC='\033[0m'

if [ -z "$1" ] && [ -z "$2" ]
then
    echo Need First Parameter \(test name to run\)
    echo Need Second Parameter \(number of runs\)
    exit 1
fi

echo Running $1, $2 times
echo

prioroutfile=''
priortracefile=''

counter=1
while [ $counter -le $2 ]
do
    outfile=\./traceruns/$1_run$counter.txt
    tracefile=\./traceruns/$1_trace$counter.txt

    echo -e ${LBLUE}Run \($counter of $2\):${NC} go test -race -cpu=1 -run $1 -trace $tracefile \> $outfile

    go test -race -cpu=1 -run $1 -trace $tracefile > $outfile
    
    if [ -z $(grep "FAIL" "$outfile") ]
    then
        echo -e ${LGREEN}"PASS"${NC}

        if [ $counter -gt 1 ]
        then
            echo -e ${DGRAY}Removing: $prioroutfile${NC}
            rm $prioroutfile
            echo -e ${DGRAY}Removing: $priortracefile${NC}
            rm $priortracefile
        fi
        
        prioroutfile=$outfile
        priortracefile=$tracefile
    else
        echo -e ${RED}"FAIL"${NC}
        mv $outfile \./traceruns/$1\_run$counter\_fail.txt
        mv $tracefile \./traceruns/$1\_trace$counter\_fail.txt
        break
    fi
    
    ((counter++))
    
    echo
done