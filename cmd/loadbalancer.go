/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math/rand"
	"net"
)

// loadbalancerCmd represents the loadbalancer command
var loadbalancerCmd = &cobra.Command{
	Use:   "loadbalancer",
	Short: "local load balancer for scaling test",
	Run: func(cmd *cobra.Command, args []string) {
		l, err := net.Listen("tcp", ":8080")
		if err != nil {
			logrus.WithError(err).Fatalln("couldn't start the load balancer")
		}
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				logrus.WithError(err).Errorln("error while accepting a new connection")
			}
			logrus.Info("accepted a new connection")
			go func() {
				defer conn.Close()
				income := make([]byte, 1024*50)
				_, err = conn.Read(income)
				if err != nil {
					logrus.WithError(err).Errorln("couldn't read the incomming message")
					return
				}
				logrus.Info("read from the client successfully")
				port := args[rand.Intn(len(args))]
				addr := "localhost:" + port
				c, err := net.Dial("tcp", addr)
				if err != nil {
					logrus.WithError(err).Errorln("couldn't connect the host")
					return
				}
				_, err = c.Write(income)
				if err != nil {
					logrus.WithError(err).Errorln("couldn't write to the host")
					return
				}
				response := make([]byte, 1024*50)
				_, err = c.Read(response)
				if err != nil {
					logrus.WithError(err).Errorln("couldn't read from the host")
					return
				}

				if _, err := conn.Write(response); err != nil {
					logrus.WithError(err).Errorln("couldn't write to the connection")
				}

				logrus.WithField("addr", addr).Info("connected successfully")

			}()
		}
	},
}

func init() {
	rootCmd.AddCommand(loadbalancerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadbalancerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadbalancerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
