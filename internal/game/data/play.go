package data

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/mokiat/ggj2024/resources"
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/util/async"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func LoadPlayData(audioAPI audio.API, engine *game.Engine, resourceSet *game.ResourceSet) async.Promise[*PlayData] {
	scenePromise := resourceSet.OpenSceneByName("World")
	airplanePromise := resourceSet.OpenModelByName("Airplane")
	ballPromise := resourceSet.OpenModelByName("Ball")
	cowPromise := resourceSet.OpenModelByName("Cow")
	burstPromise := resourceSet.OpenModelByName("Burst")
	soundtrackPromise := loadSound(audioAPI, engine, "sound/soundtrack.mp3")
	popPromise := loadSound(audioAPI, engine, "sound/pop.mp3")
	rubbingPromise := loadSound(audioAPI, engine, "sound/rubbing.mp3")
	introPromise := loadSound(audioAPI, engine, fmt.Sprintf("sound/intro-%02d.mp3", 1+random.Intn(5)))
	pilotPromise := loadSound(audioAPI, engine, fmt.Sprintf("sound/pilot-%02d.mp3", 1+random.Intn(5)))
	towerPromise := loadSound(audioAPI, engine, fmt.Sprintf("sound/tower-%02d.mp3", 1+random.Intn(4)))

	result := async.NewPromise[*PlayData]()
	go func() {
		var data PlayData
		err := errors.Join(
			scenePromise.Inject(&data.Scene),
			airplanePromise.Inject(&data.Airplane),
			ballPromise.Inject(&data.Ball),
			cowPromise.Inject(&data.Cow),
			burstPromise.Inject(&data.Burst),
			soundtrackPromise.Inject(&data.Soundtrack),
			popPromise.Inject(&data.Pop),
			rubbingPromise.Inject(&data.Rubbing),
			introPromise.Inject(&data.IntroSound),
			pilotPromise.Inject(&data.PilotSound),
			towerPromise.Inject(&data.TowerSound),
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
	Cow        *game.ModelDefinition
	Burst      *game.ModelDefinition
	Soundtrack audio.Media
	Pop        audio.Media
	Rubbing    audio.Media
	IntroSound audio.Media
	PilotSound audio.Media
	TowerSound audio.Media
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
