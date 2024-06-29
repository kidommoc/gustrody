package services

import (
	"errors"
	"reflect"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/services/auth"
	"github.com/kidommoc/gustrody/internal/services/files"
	"github.com/kidommoc/gustrody/internal/services/posts"
	"github.com/kidommoc/gustrody/internal/services/users"
)

var ErrNotPtr = errors.New("NotPointer")
var ErrNotService = errors.New("NotService")
var ErrWrongType = errors.New("WrongType")

var services = make(map[reflect.Type]interface{})

// using: Get(reflect.ValueOf(&ptr).Elem())
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
		userDbs := users.UserDbs{
			Account: userModel, Info: userModel,
			Follow: userModel, Auth: authModel,
		}
		services[ut] = users.NewService(userDbs, cfg, lg)
	}

	var pp *posts.PostService
	pt := reflect.TypeOf(pp)
	if services[pt] == nil {
		postDbs := posts.PostDbs{
			Query: postModel, Set: postModel,
			Like: postModel, Share: postModel,
		}
		us, _ := services[ut].(*users.UserService)
		services[pt] = posts.NewService(us, postDbs, cfg, lg)
	}

	var fp *files.FileService
	ft := reflect.TypeOf(fp)
	if services[ft] == nil {
		services[ft] = files.NewService(cfg, lg)
	}
}
