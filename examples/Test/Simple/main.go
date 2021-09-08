package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/platinummonkey/go-concurrency-limits/core"
	"github.com/platinummonkey/go-concurrency-limits/limit"
	"github.com/platinummonkey/go-concurrency-limits/limiter"
	"github.com/platinummonkey/go-concurrency-limits/strategy"
)

var serverC []int32
var mutex = &sync.Mutex{}
var latencyMet = 1

type transmitter struct {
	testLimiter core.Limiter
}

type loggerSet struct {
	LimitLog         *os.File
	ReqLatencyLog    *os.File
	ReqAvgLatencyLog *os.File
	SuccessNumLog    *os.File
	FailureNumLog    *os.File
	timeCosLog       *os.File
}

func (lS *loggerSet) buildLogger(ReqScale int) {
	err := os.Mkdir("examples/Test/Simple/Result", 0755)
	if err != nil {
		log.Print(err)
	}

	lS.LimitLog, err = os.OpenFile("examples/Test/Simple/Result/Scale_"+strconv.Itoa(ReqScale)+"_LimitChange"+".txt",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}

	lS.ReqLatencyLog, err = os.OpenFile("examples/Test/Simple/Result/Scale_"+strconv.Itoa(ReqScale)+"_ReqLatencyLog"+".txt",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}

	lS.ReqAvgLatencyLog, err = os.OpenFile("examples/Test/Simple/Result/Scale_"+strconv.Itoa(ReqScale)+"_ReqAvgLatencyLog"+".txt",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}

	lS.SuccessNumLog, err = os.OpenFile("examples/Test/Simple/Result/Scale_"+strconv.Itoa(ReqScale)+"_SuccessNumLog"+".txt",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}

	lS.FailureNumLog, err = os.OpenFile("examples/Test/Simple/Result/Scale_"+strconv.Itoa(ReqScale)+"_FailureNumLog"+".txt",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}

	lS.timeCosLog, err = os.OpenFile("examples/Test/Simple/Result/Scale_"+strconv.Itoa(ReqScale)+"_timeCosLog"+".txt",
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}
}

func writeSliceAtomic(latency int, latencySlice *[]int) {
	mutex.Lock()
	defer mutex.Unlock()
	*latencySlice = append(*latencySlice, latency)
}

func (tx *transmitter) transmit(ctx context.Context, timeSlot int, SucNum *int32, FailNum *int32, latencySlice *[]int, serverTest bool, AIMD bool) (bool, error) {
	id := ctx.Value(uint8(1)).(int)

	var latency time.Duration
	if latencyMet == 1 {
		latency = time.Millisecond * time.Duration(rand.Intn(10)+10*timeSlot)
	} else if latencyMet == 2 {
		latency = time.Millisecond * time.Duration(300+rand.Intn(10)-10*timeSlot)
	} else if latencyMet == 3 {
		latency = time.Millisecond * time.Duration(rand.Intn(10))
	}

	writeSliceAtomic(int(latency/time.Millisecond), latencySlice)
	token, ok := tx.testLimiter.Acquire(ctx)

	if !ok {
		atomic.AddInt32(FailNum, 1)
		if token != nil {
			token.OnDropped()
		}
		return false, fmt.Errorf("request failed for id=%d\n", id)
	}

	if AIMD {
		if int(latency/time.Millisecond) >= 20 {
			time.Sleep(time.Millisecond * 20)
			token.OnDropped()
			return false, fmt.Errorf("Time out dropped\n")
		}
	}

	if serverTest {
		//log.Println("server C", atomic.LoadInt32(&serverC[timeSlot-1]))
		if atomic.LoadInt32(&serverC[timeSlot-1]) <= 0 {
			time.Sleep(time.Millisecond * 20)
			atomic.AddInt32(FailNum, 1)
			token.OnDropped()
			return false, fmt.Errorf("request failed for id=%d\n", id)
		}
		atomic.AddInt32(&serverC[timeSlot-1], -1)
	}

	atomic.AddInt32(SucNum, 1)
	time.Sleep(latency)
	token.OnSuccess()
	//log.Printf("request succeeded for id=%d\n", id)
	return true, nil
}

func main() {
	fmt.Println("Simple Test Setting")

	ReqScalep := flag.Int("r", 100, "number of request")
	ExperimentTimeDurationp := flag.Int("eTime", 30500, "Experiment time duration")
	LimitValuep := flag.Int("lp", 100, "Limiting Value")
	ServerTestFlagp := flag.Bool("stf", true, "Server Test Flag")
	AIMDExperimentFlagp := flag.Bool("aef", true, "AIMD Experiment Flag")
	DefaultWindowSizep := flag.Int("dws", 20, "Default Window Size")
	UsingWindowp := flag.Bool("uw", true, "Using Window")
	latencyMetp := flag.Int("lm", 3, "Latency method")

	flag.Parse()

	ReqScale := *ReqScalep
	ExperimentTimeDuration := *ExperimentTimeDurationp
	LimitValue := *LimitValuep
	ServerTestFlag := *ServerTestFlagp
	AIMDExperimentFlag := *AIMDExperimentFlagp
	limiter.DefaultWindowSize = *DefaultWindowSizep
	limiter.UsingWindow = *UsingWindowp
	latencyMet = *latencyMetp

	fmt.Println(ReqScale, ExperimentTimeDuration, LimitValue,
		ServerTestFlag, AIMDExperimentFlag, limiter.DefaultWindowSize, limiter.UsingWindow, latencyMet)

	for i := 0; i < ExperimentTimeDuration/1000; i++ {
		_, cosValue := math.Sincos(float64(2) * math.Pi * float64(i) / float64(ExperimentTimeDuration/1000))
		cosValue = float64(50)*cosValue + float64(50)
		serverC = append(serverC, int32(cosValue))
	}

	limitStrategy := strategy.NewSimpleStrategy(LimitValue)
	/*testDefaultlimiter, err := limiter.NewDefaultLimiterWithDefaults(
		"Simple_Test_Limiter",
		limitStrategy,
		limitStrategy.GetLimit(),
		limit.BuiltinLimitLogger{},
		core.EmptyMetricRegistryInstance,
	)*/
	testDefaultlimiter, err := limiter.NewDefaultLimiterWithAIMD(
		"Simple_Test_Limiter",
		limitStrategy,
		limitStrategy.GetLimit(),
		0.75,
		3,
		limit.BuiltinLimitLogger{},
		core.EmptyMetricRegistryInstance,
	)
	testTransmitter := transmitter{}
	testTransmitter.testLimiter = testDefaultlimiter
	if err != nil {
		log.Fatalf("Error creating limiter err=%v\n", err)
		os.Exit(-1)
	}

	TestTotalDuration := time.NewTimer(time.Millisecond * time.Duration(ExperimentTimeDuration))
	logTicker := time.NewTicker(time.Second * 1)
	wg := sync.WaitGroup{}
	reqCounter := int32(0)
	timeSlot := 1

	lS := new(loggerSet)
	lS.buildLogger(ReqScale)
	defer lS.LimitLog.Close()
	defer lS.timeCosLog.Close()
	defer lS.FailureNumLog.Close()
	defer lS.ReqAvgLatencyLog.Close()
	defer lS.SuccessNumLog.Close()
	defer lS.ReqLatencyLog.Close()

	var SucNum int32 = 0
	var FailNum int32 = 0

	LimitLogger := log.New(lS.LimitLog, "", 0)
	//lS.ReqLatencyLog.WriteString(strconv.Itoa(timeSlot))
	perSecReqNum := ReqScale

	for {
		select {
		case <-TestTotalDuration.C:
			log.Printf("Wating for all request finished...")
			wg.Wait()
			return

		case <-logTicker.C:
			wg.Add(int(perSecReqNum))
			startTime := time.Now()
			LatencySlice := []int{}
			//log.Println("Before start server C", timeSlot, " token :", atomic.LoadInt32(&serverC[timeSlot-1]))
			lS.timeCosLog.WriteString(strconv.Itoa(timeSlot) + "	" + fmt.Sprintf("%d\n", atomic.LoadInt32(&serverC[timeSlot-1])))
			for i := 0; i < int(perSecReqNum); i++ {
				go func(j int32) {
					defer wg.Done()
					//log.Printf("%dth request is started", i)
					ctx := context.WithValue(context.Background(), uint8(1), int(j)+1)
					testTransmitter.transmit(ctx, timeSlot, &SucNum, &FailNum, &LatencySlice, ServerTestFlag, AIMDExperimentFlag)
				}(reqCounter)
				atomic.AddInt32(&reqCounter, 1)
			}
			wg.Wait()
			//log.Println("After start server C", timeSlot, " token :", atomic.LoadInt32(&serverC[timeSlot-1]))

			log.Print("Elapsed Time : ", time.Since(startTime).Seconds())
			log.Printf("limit value %d,SucNum %d %f,"+
				" FailNum %d %f, TotalNum %d, at Current slot %d time slot %d",
				limitStrategy.GetLimit(), SucNum, float64(SucNum)*100.0/float64(perSecReqNum),
				FailNum, float64(FailNum)*100.0/float64(perSecReqNum), perSecReqNum, reqCounter, timeSlot)

			LimitLogger.Printf("%d	%d", timeSlot, limitStrategy.GetLimit())
			lS.ReqLatencyLog.WriteString(strconv.Itoa(timeSlot))
			LatencyNum := len(LatencySlice)
			LatencySum := 0
			for i := 0; i < LatencyNum; i++ {
				lS.ReqLatencyLog.WriteString(" " + strconv.Itoa(LatencySlice[i]))
				LatencySum += LatencySlice[i]
			}
			lS.ReqLatencyLog.WriteString("\n")
			lS.ReqAvgLatencyLog.WriteString(strconv.Itoa(timeSlot) + "	" + fmt.Sprintf("%f\n", float64(LatencySum)/float64(LatencyNum)))
			lS.SuccessNumLog.WriteString(strconv.Itoa(timeSlot) + "	" + fmt.Sprintf("%f\n", float64(SucNum)*100.0/float64(perSecReqNum)))
			lS.FailureNumLog.WriteString(strconv.Itoa(timeSlot) + "	" + fmt.Sprintf("%f\n", float64(FailNum)*100.0/float64(perSecReqNum)))
			SucNum = 0
			FailNum = 0
			timeSlot++
			log.Println()
		}
		//reqCounter++

	}

}
