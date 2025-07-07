// Developer: Saif Hamdan
// Last Update:
// Update reason:
package handler

import (
	pb "greenlync-api-gateway/proto/vfxserver"

	// import local pkg
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/i18n"
	"greenlync-api-gateway/pkg/logger"

	"gorm.io/gorm"
	// un-comment to use model
	// model "greenlync-api-gateway/model/common"
	// repository "greenlync-api-gateway/repository"
)

var (

// your global var gose here, each with one line and

)

const (

// your const gose here , each with one line why you need it ! no unused const

)

// When other service call VFXServer they will use gPRC
// No DB instant required here

// VFXServer Server entry point
type VFXServer struct {
	// proto must have line for gPRC
	pb.UnimplementedVfxserverServer
	// Lang
	Local *i18n.Lang
	// zab logger for log to files and stdout
	Log *logger.Logger
	// Cache for Redis Caching
	Cache *cache.Cache
	// Database
	DB *gorm.DB
	// traderConfig
	// TraderConfig *traderconfig.TraderConfig
}

// Every entry will pass our main needed object from app.go, this will allow isloated runtime for each micro service

func NewVFXServer(local *i18n.Lang, l *logger.Logger, c *cache.Cache, db *gorm.DB) *VFXServer {
	// allocate memory for the objects
	o := &VFXServer{
		Local: local,
		Log:   l,
		Cache: c,
		DB:    db,
		// TraderConfig: traderConfig,
	}
	return o
}

// func (e *VFXServer) Call(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error) {
// 	// for developement
// 	//e.Lang.Tr()
// 	e.Log.Logger.Info("Received Blueprint.Call request: %v", req)
// 	// for production
// 	//e.Log.Logger.Info("Received Blueprint.Call request: %v", zapcore.Field{String: "Blueprint.Call",Interface: req})
// 	rsp := &pb.CallResponse{}
// 	rsp.Msg = "Hello " + req.Name

// 	// uncomment to use model in case we want to return model data type

// 	// mymodel = model.MyModel{
// 	// 	Id: "1",
// 	// }

// 	// DB Call , when you have DB call first check cache
// 	//value := e.Cache.Get(ctx,"key")
// 	// if value == nil {
// 	//	Get from DB
// 	//}

// 	return rsp, nil
// }
