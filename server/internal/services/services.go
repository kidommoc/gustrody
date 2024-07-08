package services

import (
	"errors"
	"reflect"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/auth"
	"github.com/kidommoc/gustrody/internal/services/federal"
	"github.com/kidommoc/gustrody/internal/services/files"
	"github.com/kidommoc/gustrody/internal/services/net"
	"github.com/kidommoc/gustrody/internal/services/posts"
	"github.com/kidommoc/gustrody/internal/services/users"
)

var ErrNotPtr = errors.New("NotPointer")
var ErrNotService = errors.New("NotService")
var ErrWrongType = errors.New("WrongType")

var services = make(map[reflect.Type]interface{})

// Using:
//
//	Get(reflect.ValueOf(&service_ptr).Elem())
func Get(v reflect.Value) error {
	t := v.Type()
	if t.Kind() != reflect.Pointer {
		return ErrNotPtr
	}
	if services[t] == nil {
		Init()
	}
	if services[t] == nil {
		return ErrNotService
	}
	x := services[t]
	if v.CanSet() && t.AssignableTo(reflect.TypeOf(x)) {
		v.Set(reflect.ValueOf(x))
		return nil
	} else {
		return ErrWrongType
	}
}

func Init() {
	cfg := config.Get()
	lg := logging.Get()
	authModel := models.AuthInstance(lg)
	userModel := models.UserInstance(lg)
	postModel := models.PostInstance(lg)

	var ap *auth.OauthService
	at := reflect.TypeOf(ap)
	if services[at] == nil {
		services[at] = auth.NewService(authModel, lg)
	}

	var up *users.UserService
	ut := reflect.TypeOf(up)
	if services[ut] == nil {
		dbs := users.UserDbs{
			Account: userModel, Info: userModel,
			Follow: userModel, Auth: authModel,
		}
		up = users.NewService(dbs, cfg, lg)
		services[ut] = up
	}

	var pp *posts.PostService
	pt := reflect.TypeOf(pp)
	if services[pt] == nil {
		dbs := posts.PostDbs{
			Query: postModel, Set: postModel,
			Like: postModel, Share: postModel,
		}
		services[pt] = posts.NewService(up, dbs, cfg, lg)
	}

	var fp *files.FileService
	ft := reflect.TypeOf(fp)
	if services[ft] == nil {
		services[ft] = files.NewService(cfg, lg)
	}

	var np *net.NetService
	nt := reflect.TypeOf(np)
	if services[nt] == nil {
		np = net.NewNetService(lg)
		services[nt] = np
	}

	var fdp *federal.FederalService
	fdt := reflect.TypeOf(fdp)
	if services[fdt] == nil {
		dbs := federal.FederalDbs{
			UserInfo: userModel, UserAccount: userModel,
			UserForeign: userModel, UserFollow: userModel,
			PostQuery: postModel, PostSet: postModel,
			PostLike: postModel, PostShare: postModel,
		}
		services[fdt] = federal.NewService(np, dbs, cfg, lg)
	}
}
