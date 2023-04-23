package main

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/pojntfx/go-nbd/pkg/server"
	"github.com/pojntfx/tapisk/pkg/backend"
	"github.com/pojntfx/tapisk/pkg/index"
	"github.com/pojntfx/tapisk/pkg/mtio"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	devFlag         = "dev"
	cacheFlag       = "cache"
	bucketFlag      = "bucket"
	sizeFlag        = "size"
	laddrFlag       = "laddr"
	networkFlag     = "network"
	nameFlag        = "name"
	descriptionFlag = "description"
	readOnlyFlag    = "read-only"
	verboseFlag     = "verbose"
)

var rootCmd = &cobra.Command{
	Use:   "tapisk",
	Short: "Expose a tape drive as a block device",
	Long: `Expose a tape drive as a block device.

Find more information at:
https://github.com/pojntfx/tapisk`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetEnvPrefix("tapisk")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		l, err := net.Listen(viper.GetString(networkFlag), viper.GetString(laddrFlag))
		if err != nil {
			return err
		}
		defer l.Close()

		log.Println("Listening on", l.Addr())

		var f *os.File
		if viper.GetBool(readOnlyFlag) {
			f, err = os.OpenFile(viper.GetString(devFlag), os.O_RDONLY, os.ModeCharDevice)
			if err != nil {
				return err
			}
		} else {
			f, err = os.OpenFile(viper.GetString(devFlag), os.O_RDWR, os.ModeCharDevice)
			if err != nil {
				return err
			}
		}
		defer f.Close()

		blocksize, err := mtio.GetBlocksize(f)
		if err != nil {
			return err
		}

		i := index.NewBboltIndex(viper.GetString(cacheFlag), viper.GetString(bucketFlag))

		if err := i.Open(); err != nil {
			return err
		}
		defer i.Close()

		b := backend.NewTapeBackend(f, i, viper.GetInt64(sizeFlag), blocksize, viper.GetBool(verboseFlag))

		errs := make(chan error)

		go func() {
			clients := 0
			for {
				conn, err := l.Accept()
				if err != nil {
					log.Println("Could not accept connection, continuing:", err)

					continue
				}

				clients++

				log.Printf("%v clients connected", clients)

				go func() {
					defer func() {
						_ = conn.Close()

						clients--

						if err := recover(); err != nil {
							log.Printf("Client disconnected with error: %v", err)
						}

						log.Printf("%v clients connected", clients)
					}()

					if err := server.Handle(
						conn,
						[]server.Export{
							{
								Name:        viper.GetString(nameFlag),
								Description: viper.GetString(descriptionFlag),
								Backend:     b,
							},
						},
						&server.Options{
							ReadOnly:           viper.GetBool(readOnlyFlag),
							MinimumBlockSize:   uint32(blocksize),
							PreferredBlockSize: uint32(blocksize),
							MaximumBlockSize:   uint32(blocksize),
						}); err != nil {
						errs <- err

						return
					}
				}()
			}
		}()

		for range errs {
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func main() {
	rootCmd.PersistentFlags().String(devFlag, "/dev/nst6", "Path to device file to connect to")
	rootCmd.PersistentFlags().String(cacheFlag, "tapisk.db", "Path to cache file to use")
	rootCmd.PersistentFlags().String(bucketFlag, "tapsik", "Bucket in cache file to use")
	rootCmd.PersistentFlags().Int64(sizeFlag, 500*1024*1024, "Size of the tape to expose (native size, not compressed size)")
	rootCmd.PersistentFlags().String(laddrFlag, ":10809", "Listen address")
	rootCmd.PersistentFlags().String(networkFlag, "tcp", "Listen network (e.g. `tcp` or `unix`)")
	rootCmd.PersistentFlags().String(nameFlag, "default", "Export name")
	rootCmd.PersistentFlags().String(descriptionFlag, "The default export", "Export description")
	rootCmd.PersistentFlags().Bool(readOnlyFlag, false, "Whether the export should be read-only")
	rootCmd.PersistentFlags().Bool(verboseFlag, false, "Whether to enable verbose logging")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(err)
	}

	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
