package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-kit/kit/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/web-zavod/svc-recognizer/pkg/service"
	"github.com/web-zavod/svc-recognizer/pkg/service/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	cfgElasticSearchURL string
	cfgGRPCPort         int
)

var rootCmd = &cobra.Command{
	Use:   "svc-search",
	Short: "Search microservice",
	Long:  "SvcSearch is a microservice that handles search queries.",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgElasticSearchURL, "elasticsearch-url", "c", "", "ElasticSearch connection URL")
	rootCmd.PersistentFlags().IntVarP(&cfgGRPCPort, "grpc-port", "g", 5030, "GRPC daemon port to connect to")

	viper.BindPFlag("elasticsearch-url", rootCmd.PersistentFlags().Lookup("elasticsearch-url"))
	viper.BindPFlag("grpc-port", rootCmd.PersistentFlags().Lookup("grpc-port"))
}

// initConfig reads in ENV variables if set.
func initConfig() {
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()
}

func run() {
	var err error

	// Create the logger
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger.Log("app", os.Args[0], "event", "starting")
	}

	// Create the Elasticsearch client
	var es *elasticsearch.Client
	{
		logger := log.With(logger, "client", "elasticsearch")

		cfg := elasticsearch.Config{
			Addresses: []string{
				viper.GetString("elasticsearch-url"),
			},
		}

		es, err = elasticsearch.NewClient(cfg)
		if err != nil {
			logger.Log("error", err)
			os.Exit(1)
		}
		logger.Log("version", elasticsearch.Version)
		logger.Log("event", "elasticsearch client successfully created")
	}

	// Create services
	var svc service.Service
	{
		logger := log.With(logger, "client", "elasticsearch")
		svc = service.NewService(es, "categories")
		if err := svc.DeleteIndex(); err != nil {
			logger.Log("error", err)
		}
		logger.Log("info", "index was deleted")

		if err := svc.CreateIndex(); err != nil {
			logger.Log("error", err)
		}
		logger.Log("info", "index was created")

		if err := svc.IndexCategory(models.Category{
			ID:       "foo",
			Category: "Продукты",
		}); err != nil {
			logger.Log("error", err)
		}
		if err := svc.IndexCategory(models.Category{
			ID:       "bar",
			Category: "Подписка",
		}); err != nil {
			logger.Log("error", err)
		}
		if err := svc.IndexCategory(models.Category{
			ID:       "baz",
			Category: "Такси",
		}); err != nil {
			logger.Log("error", err)
		}

		res, err := svc.SearchCategory("Таксо")
		if err != nil {
			logger.Log("error", err)
		}
		logger.Log("got category", res)
	}

	// Create gRPC server
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("grpc-port")))
	if err != nil {
		logger.Log("error", err)
		os.Exit(1)
	}
	srv := grpc.NewServer()

	reflection.Register(srv)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		s := <-c
		logger.Log("msg", "operating system signal received", "signal", s)

		logger.Log("msg", "waiting grpc server shut down")
		srv.GracefulStop()
	}()

	if err := srv.Serve(ln); err != nil {
		logger.Log("error", err)
		os.Exit(1)
	}
	logger.Log("msg", "grpc server stopped normally")
}
