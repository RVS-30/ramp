package analyser

import "encoding/xml"

var javaFrameworkMarkers = map[string]string{
	"spring-boot-starter": "Spring Boot",
	"spring-core":         "Spring",
	"spring-webmvc":       "Spring MVC",
	"quarkus":             "Quarkus",
	"micronaut":           "Micronaut",
	"dropwizard":          "Dropwizard",
	"vertx":               "Vert.x",
	"hibernate-core":      "Hibernate",
}

var javaFrameworkPriority = []string{
	"Spring Boot", "Spring MVC", "Spring", "Quarkus", "Micronaut", "Dropwizard", "Vert.x",
	"Hibernate",
}

var javaToolingMarkers = map[string]string{
	"junit":   "JUnit",
	"testng":  "TestNG",
	"mockito": "Mockito",
}

var javaToolingPriority = []string{
	"JUnit", "TestNG", "Mockito",
}

type pomXML struct {
	XMLName    xml.Name `xml:"project"`
	ArtifactID string   `xml:"artifactId"`
	Properties struct {
		JavaVersion string `xml:"java.version"`
	} `xml:"properties"`
	Dependencies struct {
		Dependency []struct {
			ArtifactID string `xml:"artifactId"`
		} `xml:"dependency"`
	} `xml:"dependencies"`
}

func parsePomXML(data []byte, info *ProjectInfo) {
	var pom pomXML
	if err := xml.Unmarshal(data, &pom); err != nil {
		return // malformed pom.xml — leave info as-is, caller falls back
	}

	if pom.ArtifactID != "" {
		info.Name = pom.ArtifactID
	}
	if pom.Properties.JavaVersion != "" {
		info.Version = "Java " + pom.Properties.JavaVersion
	}

	deps := make(map[string]bool)
	for _, dep := range pom.Dependencies.Dependency {
		deps[dep.ArtifactID] = true
	}

	if matchFramework(deps, javaFrameworkMarkers, javaFrameworkPriority, info) {
		return
	}
	matchFramework(deps, javaToolingMarkers, javaToolingPriority, info)
}