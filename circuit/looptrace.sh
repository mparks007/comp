RED='\033[0;31m'
LGREEN='\033[1;32m'
LBLUE='\033[1;34m'
DGRAY='\033[1;30m'
NC='\033[0m'

# gentle reminder of params, only if NO params supplied (nothing fancy pursued here)
if [ -z "$1" ] && [ -z "$2" ]
then
    echo Need First Parameter \(test name to run\)
    echo Need Second Parameter \(number of runs\)
    echo
    echo [Optional] 'r' as Third Parameter if wanting race detection
    exit 1
fi

echo Running $1, $2 times
echo

# want to use race detection?
race=''
if [[ $3 = 'r' ]]
then
    race='-race'
fi

prioroutfile=''
priortracefile=''

counter=1
while [ $counter -le $2 ]
do
    outfile=\./traceruns/$1_run$counter.txt
    tracefile=\./traceruns/$1_trace$counter.txt

    echo -e ${LBLUE}Run \($counter of $2\):${NC} go test $race -cpu=1 -run $1 -debug -trace $tracefile \> $outfile

    go test $race -cpu=1 -run $1 -debug -trace $tracefile > $outfile

    # basic check for known text in go test fail message (tighten up if other debug output would have FAIL in it)
    if [[ -z $(grep 'FAIL' $outfile) ]]
    then
        echo -e ${LGREEN}"PASS"${NC}

        # clean up PRIOR run if was a Pass, leaving just the new Pass files until next time around (ensures no huge list of files for passes, but at least one recent set)
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

        # rename to avoid overwrite if run this whole script again and test fails on same iteration count
        mv $outfile \./traceruns/$1\_run$counter\_fail.txt
        mv $tracefile \./traceruns/$1\_trace$counter\_fail.txt
        break
    fi
    
    ((counter++))
    
    echo
done