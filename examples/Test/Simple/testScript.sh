#!/usr/bin/env bash

: <<'END'
ReqScalep := flag.Int("r", 100, "number of request")
ExperimentTimeDurationp := flag.Int("eTime", 30500, "Experiment time duration")
LimitValuep := flag.Int("lp", 100, "Limiting Value")
ServerTestFlagp := flag.Bool("stf", false, "Server Test Flag")
AIMDExperimentFlagp := flag.Bool("aef", false, "AIMD Experiment Flag")
DefaultWindowSizep := flag.Int("dws", 200, "Default Window Size")
UsingWindowp := flag.Bool("uw", true, "Using Window")
latencyMetp := flag.Int("lm", 1, "Latency method")
END

for win in 1 20 100 200
do
  for lm in 1 2 3
  do
    if [ ${win} -eq 1 ];then
        echo "No window case"
        /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/main "-r"=100 "-eTime"=30500 "-lp"=100 "-stf"=false "-aef"=true "-dws"=10 "-uw"=false "-lm"=$lm \
        1> /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/result/"$win"_"$lm".txt \
        2>&1
    else
        echo "Yes window case"
        /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/main "-r"=100 "-eTime"=30500 "-lp"=100 "-stf"=false "-aef"=true "-dws"=$win "-uw"=true "-lm"=$lm \
        1> /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/result/"$win"_"$lm".txt \
        2>&1
    fi

    mkdir /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/result/"$win"_"$lm"
    wait
    echo "finish and move"
    mv /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/result/*.txt /Users/user/IdeaProjects/go-concurrency-limits/examples/Test/Simple/result/"$win"_"$lm"
  done
done




