package main

import (
	"fmt"
	"log"

	"github.com/mokiat/lacking/data/pack"
	"github.com/mokiat/lacking/game/asset"
)

func main() {
	if err := runTool(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runTool() error {
	registry, err := asset.NewDirRegistry(".")
	if err != nil {
		return fmt.Errorf("failed to create registry: %w", err)
	}

	skybox := ensureResource(registry, "414eb145-3c3b-4d90-ad53-7aede66bc9c1", "cube_texture", "Skybox")
	skyboxReflection := ensureResource(registry, "ba7fb3b4-20c3-44f6-89c0-e6e34607209f", "cube_texture", "Skybox Reflection")
	skyboxRefraction := ensureResource(registry, "c233eac3-96fb-4b40-88c0-5e7f6bf564e1", "cube_texture", "Skybox Refraction")

	modelWorld := ensureResource(registry, "90345c66-2194-4a2e-acea-04b21c2df048", "model", "World")
	modelAirplane := ensureResource(registry, "41b37fbc-5428-477b-8c7a-8bb58ac34514", "model", "Airplane")
	modelBall := ensureResource(registry, "61cbde74-436e-4306-b3cc-b0c2459dbecb", "model", "Ball")
	modelCow := ensureResource(registry, "4d6c54e9-9152-4c35-8f33-8fd9f898b091", "model", "Cow")
	modelBurst := ensureResource(registry, "988992d4-2661-468a-baf3-298b1f6764d7", "model", "Burst")

	levelWorld := ensureResource(registry, "21a3cecd-6d04-4fcf-9c9d-e210b97dad3f", "scene", "World")
	levelWorld.AddDependency(skybox)
	levelWorld.AddDependency(skyboxReflection)
	levelWorld.AddDependency(skyboxRefraction)
	levelWorld.AddDependency(modelWorld)
	levelWorld.AddDependency(modelAirplane)
	levelWorld.AddDependency(modelBall)
	levelWorld.AddDependency(modelBurst)

	if err := registry.Save(); err != nil {
		return fmt.Errorf("error saving resources: %w", err)
	}

	packer := pack.NewPacker(registry)

	// Cube Textures
	packer.Pipeline(func(p *pack.Pipeline) {
		equirectangularImage := p.OpenImageResource("resources/images/skybox.hdr")
		cubeImage := p.BuildCubeImage(
			pack.WithFrontImage(p.BuildCubeSideFromEquirectangular(pack.CubeSideFront, equirectangularImage)),
			pack.WithRearImage(p.BuildCubeSideFromEquirectangular(pack.CubeSideRear, equirectangularImage)),
			pack.WithLeftImage(p.BuildCubeSideFromEquirectangular(pack.CubeSideLeft, equirectangularImage)),
			pack.WithRightImage(p.BuildCubeSideFromEquirectangular(pack.CubeSideRight, equirectangularImage)),
			pack.WithTopImage(p.BuildCubeSideFromEquirectangular(pack.CubeSideTop, equirectangularImage)),
			pack.WithBottomImage(p.BuildCubeSideFromEquirectangular(pack.CubeSideBottom, equirectangularImage)),
		)

		smallerCubeImage := p.ScaleCubeImage(cubeImage, 512)
		p.SaveCubeTextureAsset(skybox, smallerCubeImage,
			pack.WithFormat(asset.TexelFormatRGBA16F),
		)

		reflectionCubeImage := p.ScaleCubeImage(cubeImage, 128)
		p.SaveCubeTextureAsset(skyboxReflection, reflectionCubeImage,
			pack.WithFormat(asset.TexelFormatRGBA16F),
		)

		refractionCubeImage := p.BuildIrradianceCubeImage(reflectionCubeImage,
			pack.WithSampleCount(50),
		)
		p.SaveCubeTextureAsset(skyboxRefraction, refractionCubeImage,
			pack.WithFormat(asset.TexelFormatRGBA16F),
		)
	})

	// Models
	packer.Pipeline(func(p *pack.Pipeline) {
		p.SaveModelAsset(modelWorld,
			p.OpenGLTFResource("resources/models/scene.glb"),
		)

		p.SaveModelAsset(modelAirplane,
			p.OpenGLTFResource("resources/models/airplane.glb"),
		)

		p.SaveModelAsset(modelBall,
			p.OpenGLTFResource("resources/models/ball.glb"),
		)

		p.SaveModelAsset(modelCow,
			p.OpenGLTFResource("resources/models/cow.glb"),
		)

		p.SaveModelAsset(modelBurst,
			p.OpenGLTFResource("resources/models/burst.glb"),
		)
	})

	// Levels
	packer.Pipeline(func(p *pack.Pipeline) {
		p.SaveLevelAsset(levelWorld,
			p.OpenLevelResource("resources/levels/world.json"),
		)
	})

	return packer.RunParallel()
}

func ensureResource(registry asset.Registry, id, kind, name string) asset.Resource {
	resource := registry.ResourceByID(id)
	if resource == nil {
		resource = registry.CreateIDResource(id, kind, name)
	} else {
		resource.SetName(name)
	}
	return resource
}
