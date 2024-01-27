package data

import (
	"errors"
	"io"

	"github.com/mokiat/ggj2024/resources"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/util/async"
)

func LoadPlayData(audioAPI audio.API, engine *game.Engine, resourceSet *game.ResourceSet) async.Promise[*PlayData] {
	scenePromise := resourceSet.OpenSceneByName("World")
	airplanePromise := resourceSet.OpenModelByName("Airplane")
	ballPromise := resourceSet.OpenModelByName("Ball")
	soundPromise := loadSound(audioAPI, engine, "sound/soundtrack.mp3")

	result := async.NewPromise[*PlayData]()
	go func() {
		var data PlayData
		err := errors.Join(
			scenePromise.Inject(&data.Scene),
			airplanePromise.Inject(&data.Airplane),
			ballPromise.Inject(&data.Ball),
			soundPromise.Inject(&data.Soundtrack),
		)
		if err != nil {
			result.Fail(err)
		} else {
			result.Deliver(&data)
		}
	}()
	return result
}

type PlayData struct {
	Scene      *game.SceneDefinition
	Airplane   *game.ModelDefinition
	Ball       *game.ModelDefinition
	Soundtrack audio.Media
}

func loadSound(audioAPI audio.API, engine *game.Engine, name string) async.Promise[audio.Media] {
	result := async.NewPromise[audio.Media]()

	go func() {
		file, err := resources.Sound.Open(name)
		if err != nil {
			result.Fail(err)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			result.Fail(err)
			return
		}

		var media audio.Media
		engine.IOWorker().Schedule(func() error {
			media = audioAPI.CreateMedia(audio.MediaInfo{
				Data:     data,
				DataType: audio.MediaDataTypeAuto,
			})
			return nil
		}).Wait()
		result.Deliver(media)
	}()

	return result
}
