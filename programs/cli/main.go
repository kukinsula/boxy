package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"
	loginEntity "github.com/kukinsula/boxy/entity/login"
	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
	"github.com/kukinsula/boxy/framework/api/client"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/linechart"
)

func main() {
	gui()
	// stress()
}

func gui() {
	logger := log.CleanMetaLogger(log.StdoutLogger)
	service := client.NewService("http://127.0.0.1:9000",
		func(req *client.Request) {
			logger(req.UUID, log.DEBUG,
				fmt.Sprintf("HTTP -> %s %s", req.Method, req.URL),
				map[string]interface{}{
					"headers": req.Headers,
					"body":    req.Body,
				})
		},

		func(resp *client.Response) {
			logger(resp.Request.UUID, log.DEBUG,
				fmt.Sprintf("HTTP <- %s %s", resp.Request.Method, resp.Request.URL),
				map[string]interface{}{
					"duration": resp.Duration,
					"headers":  resp.Headers,
					"status":   resp.Status,
					"error":    resp.Error,
				})
		})

	signinResult, err := service.Login.Signin(entity.NewUUID(), &loginUsecase.SigninParams{
		Email:    "fhBb.rykEy@mail.io",
		Password: "LswSuyjg",
	})

	if err != nil {
		panic(err)
	}

	channel := make(chan *monitoringEntity.Metrics)
	err = service.Streaming.Stream(entity.NewUUID(), signinResult.AccessToken, channel)
	if err != nil {
		panic(fmt.Sprintf("Streaming failed: %v", err))
	}

	// GUI

	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	cpu, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorBlue,
			cell.ColorRed,
			cell.ColorYellow,
			cell.ColorBlue,
			cell.ColorGreen,
		}),

		barchart.ValueColors([]cell.Color{
			cell.ColorRed,
			cell.ColorYellow,
			cell.ColorBlue,
			cell.ColorGreen,
			cell.ColorRed,
		}),

		barchart.ShowValues(),
		barchart.BarWidth(5),
		barchart.Labels([]string{
			"core0",
			"core1",
			"core2",
			"core3",
			"cpu",
		}),
	)

	if err != nil {
		panic(err)
	}

	memOccupied, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("Memory occupied", cell.FgColor(cell.ColorGreen)),
	)

	swapOccupied, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("SWAP occupied", cell.FgColor(cell.ColorYellow)),
	)

	vmAllocOccupied, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("Virual Memory occupied", cell.FgColor(cell.ColorBlue)),
	)

	// NETWORK

	network, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	)
	if err != nil {
		panic(err)
	}

	go play(ctx,
		cpu,
		memOccupied, swapOccupied, vmAllocOccupied,
		network,
		channel)

	c, err := container.New(t,
		container.Border(linestyle.Light),
		container.BorderTitle("MONITORING"),

		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(container.PlaceWidget(cpu)),
					container.Right(
						container.SplitHorizontal(
							container.Top(
								container.SplitVertical(
									container.Left(container.PlaceWidget(memOccupied)),
									container.Right(container.PlaceWidget(swapOccupied)),
								),
							),
							container.Bottom(container.PlaceWidget(vmAllocOccupied)),
						),
					),
				),
			),
			container.Bottom(container.PlaceWidget(network)),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	err = termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter))
	if err != nil {
		panic(err)
	}
}

func play(
	ctx context.Context,
	cpu *barchart.BarChart,
	memOccupied *donut.Donut,
	swapOccupied *donut.Donut,
	vmAllocOccupied *donut.Donut,
	network *linechart.LineChart,
	channel chan *monitoringEntity.Metrics) {

	downloads := make([]float64, 10)
	uploads := make([]float64, 10)

	for {
		select {
		case metrics := <-channel:
			// CPU

			values := make([]int, len(metrics.CPU.LoadAverages)+1)

			for index := 0; index < len(metrics.CPU.LoadAverages); index++ {
				values[index] = int(metrics.CPU.LoadAverages[index])
			}

			values[len(metrics.CPU.LoadAverages)] = int(metrics.CPU.LoadAverage)

			err := cpu.Values(values, 100)
			if err != nil {
				panic(err)
			}

			// MEMORY

			percent := metrics.Memory.CurrentMeasure.MemOccupied * 100 / metrics.Memory.CurrentMeasure.MemTotal
			err = memOccupied.Percent(int(percent))
			if err != nil {
				panic(err)
			}

			percent = metrics.Memory.CurrentMeasure.SwapOccupied * 100 / metrics.Memory.CurrentMeasure.SwapTotal
			err = swapOccupied.Percent(int(percent))
			if err != nil {
				panic(err)
			}

			percent = metrics.Memory.CurrentMeasure.VmallocOccupied * 100 / metrics.Memory.CurrentMeasure.VmallocTotal
			err = vmAllocOccupied.Percent(int(percent))
			if err != nil {
				panic(err)
			}

			// NETWORK

			for _, net := range metrics.Network.CurrentMeasure {
				downloads = append(downloads, net.Download)
				uploads = append(downloads, net.Upload)

				if len(downloads) == 100 {
					downloads = downloads[1:]
				}

				if len(uploads) == 100 {
					uploads = uploads[1:]
				}
			}

			err = network.Series("Download", downloads,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorRed)),
				linechart.SeriesXLabels(map[int]string{0: "0"}),
			)

			if err != nil {
				panic(err)
			}

			err = network.Series("Upload", uploads,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
				linechart.SeriesXLabels(map[int]string{0: "0"}),
			)

			if err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func stress() {
	logger := log.CleanMetaLogger(log.StdoutLogger)

	service := client.NewService("http://127.0.0.1:9000",
		func(req *client.Request) {
			logger(req.UUID, log.DEBUG,
				fmt.Sprintf("HTTP -> %s %s", req.Method, req.URL),
				map[string]interface{}{
					"headers": req.Headers,
					"body":    req.Body,
				})
		},

		func(resp *client.Response) {
			logger(resp.Request.UUID, log.DEBUG,
				fmt.Sprintf("HTTP <- %s %s", resp.Request.Method, resp.Request.URL),
				map[string]interface{}{
					"duration": resp.Duration,
					"headers":  resp.Headers,
					"status":   resp.Status,
					"error":    resp.Error,
				})
		})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	rand.Seed(time.Now().UnixNano())

	go func() {
		workers := 1

		for index := 0; index < workers; index++ {
			user := &loginEntity.User{
				Password:  GenerateRandomString(8),
				FirstName: GenerateRandomString(4),
				LastName:  GenerateRandomString(5),
			}

			user.Email = fmt.Sprintf("%s.%s@mail.io", user.FirstName, user.LastName)

			go worker(service, logger, user, 2000, 5000)
		}
	}()

	<-signals

	logger(entity.NewUUID(), log.INFO, "Finished!", nil)
}

func worker(
	service *client.Service,
	logger log.Logger,
	user *loginEntity.User,
	min, max int) {

	var signinResult *loginUsecase.SigninResult
	var meResult *loginUsecase.SigninResult
	// var streamer *client.Streamer
	var err error

	randomer := rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(getRandomDuration(randomer, min, max))

	signupResult, err := service.Login.Signup(entity.NewUUID(), &loginUsecase.CreateUserParams{
		Email:     user.Email,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})

	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "worker Signup failed",
			map[string]interface{}{"error": err})
		return
	}

	err = service.Login.CheckActivate(entity.NewUUID(), &loginUsecase.EmailAndTokenParams{
		Email: signupResult.Email,
		Token: signupResult.ActivationToken,
	})

	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "worker CheckActivate failed",
			map[string]interface{}{"error": err})
		return
	}

	err = service.Login.Activate(entity.NewUUID(), &loginUsecase.EmailAndTokenParams{
		Email: signupResult.Email,
		Token: signupResult.ActivationToken,
	})

	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "worker Activate failed",
			map[string]interface{}{"error": err})
		return
	}

	for err == nil {
		time.Sleep(getRandomDuration(randomer, min, max))

		signinResult, err = service.Login.Signin(entity.NewUUID(), &loginUsecase.SigninParams{
			Email:    user.Email,
			Password: user.Password,
		})

		if err != nil {
			break
		}

		time.Sleep(getRandomDuration(randomer, min, max))

		meResult, err = service.Login.Me(entity.NewUUID(), signinResult.AccessToken)
		if err != nil {
			break
		}

		time.Sleep(getRandomDuration(randomer, min, max))

		channel := make(chan *monitoringEntity.Metrics)
		err = service.Streaming.Stream(entity.NewUUID(), meResult.AccessToken, channel)
		if err != nil {
			break
		}

		for index := 0; index < getRandom(randomer, 5, 20); index++ {
			metrics := <-channel

			logger(entity.NewUUID(), log.DEBUG, "METRICS",
				map[string]interface{}{
					"CPU":     metrics.CPU,
					"Memory":  metrics.Memory,
					"Network": metrics.Network,
				})

			fmt.Println(index, metrics)
		}

		time.Sleep(getRandomDuration(randomer, min, max))

		err = service.Login.Logout(entity.NewUUID(), meResult.AccessToken)
		if err != nil {
			break
		}
	}

	logger(entity.NewUUID(), log.ERROR, "worker ended",
		map[string]interface{}{"error": err})
}

func Stream(service *client.Service, token string, logger log.Logger) {
	channel := make(chan *monitoringEntity.Metrics)
	err := service.Streaming.Stream(entity.NewUUID(), token, channel)
	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "streaming failed",
			map[string]interface{}{"error": err})
	}

	for metrics := range channel {
		fmt.Printf("METRICS %s\n", metrics)
	}
}

func getRandom(randomer *rand.Rand, min, max int) int {
	return rand.Intn(max-min) + min
}

func getRandomDuration(randomer *rand.Rand, min, max int) time.Duration {
	return time.Duration(getRandom(randomer, min, max)) * time.Millisecond
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
