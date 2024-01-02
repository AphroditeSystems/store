package main

import (
	"context"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "github.com/AphroditeSystems/store/proto"
)

type server struct {
	pb.UnimplementedStoreServiceServer
}

type Media struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Filename  string
	Extension string
	Type      string //TODO: maybe enum?
}

var DB *gorm.DB

func (s *server) StoreMedia(ctx context.Context, req *pb.StoreMediaRequest) (*pb.StoreMediaResponse, error) {
	log.Info().Msgf("received file: %v.%v", req.Filename, req.Extension)

	var media_type string

	switch req.Extension {
	case "jpg":
		media_type = "image"
	case "png":
		media_type = "image"
	case "webp":
		media_type = "image"
	case "mp4":
		media_type = "video"
	default:
		media_type = "unknown"
	}

	if DB == nil {
		log.Fatal().Msg("DB is nil")
	}

	media := Media{
		Filename:  req.Filename,
		Extension: req.Extension,
		Type:      media_type,
	}

	DB.Create(&media)

	// Create ./data directory if it doesn't exist
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		os.Mkdir("./data", os.ModeDir)
	}

	// store file in data with uuid as filename
	file, err := os.Create("./data/" + media.ID.String())
	if err != nil {
		log.Error().Err(err).Msg("failed to create file")
		return nil, err
	}
	defer file.Close()

	_, err = file.Write(req.Data)
	if err != nil {
		log.Error().Err(err).Msg("failed to write file")
		return nil, err
	}

	return &pb.StoreMediaResponse{}, nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	// TODO: Move to config file
	dsn := "host=localhost user=as password=changeme dbname=as port=5432"
	// Suppress gorm logs
	var err error
	DB, err = gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			// Logger: logger.Default.LogMode(logger.Silent)
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	DB.AutoMigrate(&Media{})

	listener, err := net.Listen("tcp", "localhost:9998")
	if err != nil {
		panic(err)
	}
	// TODO: Serve files over HTTP

	s := grpc.NewServer()
	pb.RegisterStoreServiceServer(s, &server{})

	log.Info().Msg("server started on port 9998")

	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
