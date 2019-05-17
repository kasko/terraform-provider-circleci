package circleci

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccCircleCIProject_basic(t *testing.T) {
	var proj Project

	org := os.Getenv("CIRCLECI_TEST_ORGANIZATION")
	repo := os.Getenv("CIRCLECI_TEST_REPO")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCircleCIProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCircleCIProject_basic(org, repo),
				Check: resource.ComposeTestCheckFunc(
					testCheckCircleCIProjectExists("circleci_project.project", &proj),
					resource.TestCheckResourceAttr("circleci_project.project", "vcs_type", "github"),
					resource.TestCheckResourceAttr("circleci_project.project", "account", org),
					resource.TestCheckResourceAttr("circleci_project.project", "project", repo),
					resource.TestCheckResourceAttr("circleci_project.project", "variable.3244239841.name", "__________X_FOO"),
					resource.TestCheckResourceAttr("circleci_project.project", "variable.3244239841.value", "xxxxr"),
					testAccCheckCircleCiProjectAttributes(&proj, &testAccCircleCIProjectExpectedAttributes{}),
				),
			},
			{
				Config: testAccCircleCIProject_basicUpdated(org, repo),
				Check: resource.ComposeTestCheckFunc(
					testCheckCircleCIProjectExists("circleci_project.project", &proj),
					resource.TestCheckResourceAttr("circleci_project.project", "vcs_type", "github"),
					resource.TestCheckResourceAttr("circleci_project.project", "account", org),
					resource.TestCheckResourceAttr("circleci_project.project", "project", repo),
					resource.TestCheckResourceAttr("circleci_project.project", "variable.1444073175.name", "RENAMED_X_FOO"),
					resource.TestCheckResourceAttr("circleci_project.project", "variable.1444073175.value", "xxxxr"),
					resource.TestCheckResourceAttr("circleci_project.project", "variable.3732134584.name", "X_FIZZ"),
					resource.TestCheckResourceAttr("circleci_project.project", "variable.3732134584.value", "xxxxzz"),
					resource.TestCheckNoResourceAttr("circleci_project.project", "variable.3244239841.name"),
					resource.TestCheckNoResourceAttr("circleci_project.project", "variable.3244239841.value"),
					testAccCheckCircleCiProjectAttributes(&proj, &testAccCircleCIProjectExpectedAttributes{}),
				),
			},
		},
	})
}

func testCheckCircleCIProjectExists(n string, proj *Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CircleCI Project ID is set.")
		}

		conn := testAccProvider.Meta().(*ApiClient)

		gotProj, err := conn.GetProject("github", rs.Primary.Attributes["account"], rs.Primary.Attributes["project"])

		if err != nil {
			return err
		}

		*proj = *gotProj

		return nil
	}
}

type testAccCircleCIProjectExpectedAttributes struct {
}

func testAccCheckCircleCiProjectAttributes(proj *Project, want *testAccCircleCIProjectExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Some extra checks on Project (returned via API) can be done here.

		// if true == false {
		// 	return fmt.Errorf("error message")
		// }

		return nil
	}
}

func testCheckCircleCIProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "circleci_project" {
			continue
		}

		_, err := conn.GetProject("github", rs.Primary.Attributes["account"], rs.Primary.Attributes["project"])

		if err == nil {
			return fmt.Errorf("Expected CircleCi project to be gone, but was still found.")
		}

		return nil
	}

	return fmt.Errorf("Default error in CircleCI Project Test")
}

func testAccCircleCIProject_basic(org, repo string) string {
	return fmt.Sprintf(`
resource "circleci_project" "project" {
  vcs_type = "github"
  account  = "%s"
  project  = "%s"

  variable {
    name  = "__________X_FOO"
    value = "bar"
  }
}
`, org, repo)
}

func testAccCircleCIProject_basicUpdated(org, repo string) string {
	return fmt.Sprintf(`
resource "circleci_project" "project" {
  vcs_type = "github"
  account  = "%s"
  project  = "%s"

  variable {
    name  = "RENAMED_X_FOO"
    value = "bar"
  }

  variable {
    name  = "X_FIZZ"
    value = "buzz"
  }
}
`, org, repo)
}
