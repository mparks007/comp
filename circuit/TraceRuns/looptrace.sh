if [ -z "$1" ] && [ -z "$2" ]
then
    echo Need First Parameter \(test name to run\)
    echo Need Second Parameter \(number of runs\)
    exit 1
fi

echo Running $1, $2 times

counter=1
while [ $counter -le $2 ]
do
    echo $counter
    go test ../ -race -cpu=1 -run $1 -trace trace$counter.txt > run$counter.txt
    ((counter++))
done