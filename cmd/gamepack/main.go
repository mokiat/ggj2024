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

	skybox := ensureResource(registry, "bab99e80-ded1-459a-b00b-6a17afa44046", "cube_texture", "Skybox")
	skyboxReflection := ensureResource(registry, "eb639f55-d6eb-46d7-bd3b-d52fcaa0bc58", "cube_texture", "Skybox Reflection")
	skyboxRefraction := ensureResource(registry, "0815fb89-7ae6-4229-b9e2-59610c4fc6bc", "cube_texture", "Skybox Refraction")

	modelWorld := ensureResource(registry, "5f7bd967-dc4a-4252-b1a5-5721cd299d67", "model", "World")

	levelWorld := ensureResource(registry, "884e6395-2300-47bb-9916-b80e3dc0e086", "scene", "World")
	levelWorld.AddDependency(skybox)
	levelWorld.AddDependency(skyboxReflection)
	levelWorld.AddDependency(skyboxRefraction)
	levelWorld.AddDependency(modelWorld)

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
