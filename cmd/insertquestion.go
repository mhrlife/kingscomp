/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"kingscomp/internal/entity"
	"kingscomp/internal/repository"
	"kingscomp/internal/repository/redis"
	"kingscomp/pkg/jsonhelper"
	"os"
)

// insertQuestionCmd represents the insertQuestion command
var insertQuestionCmd = &cobra.Command{
	Use:   "insertquestion",
	Short: "insert a list of questions",
	Run: func(cmd *cobra.Command, args []string) {

		json, _ := cmd.Flags().GetString("json")

		if json == "" {
			logrus.Fatalln("please enter the file-path using --file-path")
		}

		questions := jsonhelper.Decode[[]entity.Question]([]byte(json))

		_ = godotenv.Load()
		// set up repositories
		redisClient, err := redis.NewRedisClient(os.Getenv("REDIS_URL"))
		if err != nil {
			logrus.WithError(err).Fatalln("couldn't connect to te redis server")
		}
		questionRepository := repository.NewQuestionRedisRepository(redisClient)
		_ = questionRepository

		logrus.WithField("num", len(questions)).Info("inserting new questions")
		err = questionRepository.PushActiveQuestion(context.Background(), questions...)
		if err != nil {
			logrus.WithError(err).Fatalln("couldn't push the new questions")
		}

		logrus.Info("questions have been added successfully")

	},
}

func init() {
	rootCmd.AddCommand(insertQuestionCmd)
	insertQuestionCmd.PersistentFlags().String("json", "", "path of the JSON questions file")
}
