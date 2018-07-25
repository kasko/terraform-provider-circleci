package circleci

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "This is the GitHub or Bitbucket project account (organization) name for the target project (not your personal GitHub or Bitbucket username).",
			},
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "This is the GitHub or Bitbucket project (repository) name.",
			},
			"vcs_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "github",
				Description: "Version control system type your project uses.",
				ValidateFunc: func(v interface{}, k string) (ws []string, errs []error) {
					value := v.(string)
					if value != "github" && value != "bitbucket" {
						errs = append(errs, fmt.Errorf("Value of vcs_type must be either github or bitbucket."))
					}
					return
				},
			},
			"aws_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"keypair": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "AWS access key id.",
										StateFunc: func(v interface{}) string {
											return maskCircleCiSecret(v.(string))
										},
									},
									"secret_key": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "AWS secret key.",
										StateFunc: func(v interface{}) string {
											return maskCircleCiSecret(v.(string))
										},
									},
								},
							},
						},
					},
				},
			},
			"variable": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(v interface{}) string {
								return maskCircleCiSecret(v.(string))
							},
						},
					},
				},
				Set: variableHash,
			},
		},
	}
}

func resourceProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ApiClient)

	vcstype := d.Get("vcs_type").(string)
	account := d.Get("account").(string)
	reponame := d.Get("project").(string)

	log.Printf("[DEBUG] Following %s/%s %s project on CircleCI", account, reponame, vcstype)

	d.Partial(true)

	_, err := client.FollowProject(vcstype, account, reponame)
	if err != nil {
		return fmt.Errorf("error following project: %s", err)
	}

	d.SetId(buildId(vcstype, account, reponame))

	if _, ok := d.GetOk("aws_config.0.keypair"); ok {
		err := client.SetAwsKeys(
			vcstype,
			account,
			reponame,
			d.Get("aws_config.0.keypair.0.access_key").(string),
			d.Get("aws_config.0.keypair.0.secret_key").(string),
		)

		if err != nil {
			return err
		}

		d.SetPartial("aws_config.0.keypair")
	}

	d.Partial(false)

	return resourceProjectUpdate(d, meta)
}

func resourceProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ApiClient)

	vcstype, account, reponame := expandId(d.Id())

	project, err := client.GetProject(vcstype, account, reponame)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error reading CircleCI project %q: %s", d.Id(), err)
	}

	d.Set("vcs_type", project.VcsType)
	d.Set("account", project.Username)
	d.Set("project", project.Reponame)

	if err := flattenAwsConfig(d, project); err != nil {
		return fmt.Errorf("Error setting aws_config: %v", err)
	}

	envVars, err := client.ListEnvVars(vcstype, account, reponame)

	if err := flattenEnvironmentVariables(d, envVars); err != nil {
		return fmt.Errorf("Error setting environment: %v", err)
	}

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ApiClient)

	vcstype, account, reponame := expandId(d.Id())

	d.Partial(true)

	if d.HasChange("aws_config.0.keypair") {
		oraw, nraw := d.GetChange("aws_config.0.keypair")
		o := oraw.([]interface{})
		n := nraw.([]interface{})

		if len(o) > 0 && len(n) == 1 {
			err := client.SetAwsKeys(
				vcstype,
				account,
				reponame,
				d.Get("aws_config.0.keypair.0.access_key").(string),
				d.Get("aws_config.0.keypair.0.secret_key").(string),
			)

			if err != nil {
				return err
			}
		}

		if len(n) == 0 && len(o) == 1 {
			err := client.RemoveAwsKeys(
				vcstype,
				account,
				reponame,
			)

			if err != nil {
				return err
			}
		}

		d.SetPartial("aws_config.0.keypair")
	}

	if d.HasChange("variable") {
		o, n := d.GetChange("variable")

		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		for _, pRaw := range ns.Difference(os).List() {
			data := pRaw.(map[string]interface{})

			_, err := client.AddEnvVar(
				vcstype,
				account,
				reponame,
				data["name"].(string),
				data["value"].(string),
			)

			if err != nil {
				return err
			}
		}

		for _, pRaw := range os.Difference(ns).List() {
			data := pRaw.(map[string]interface{})

			err := client.DeleteEnvVar(
				vcstype,
				account,
				reponame,
				data["name"].(string),
			)

			if err != nil {
				return err
			}
		}

		d.SetPartial("variable")
	}

	d.Partial(false)

	return resourceProjectRead(d, meta)
}

func resourceProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ApiClient)

	vcstype, account, reponame := expandId(d.Id())

	err := client.DisableProject(vcstype, account, reponame)
	if err != nil {
		return fmt.Errorf("Error disabling project %q: %s", d.Id(), err)
	}

	return nil
}

func flattenAwsConfig(d *schema.ResourceData, project *Project) error {
	awsConfig := project.AWSConfig

	conf := map[string]interface{}{}

	if awsConfig.AWSKeypair != nil {
		conf["keypair"] = []interface{}{
			map[string]interface{}{
				"access_key": awsConfig.AWSKeypair.AccessKey,
				"secret_key": awsConfig.AWSKeypair.SecretKey,
			},
		}
	}

	if err := d.Set("aws_config", []interface{}{
		conf,
	}); err != nil {
		return err
	}

	return nil
}

func flattenEnvironmentVariables(d *schema.ResourceData, vars []EnvVar) error {
	variables := make([]map[string]interface{}, 0, len(vars))

	for _, v := range vars {
		variable := make(map[string]interface{})

		variable["name"] = v.Name
		variable["value"] = v.Value

		variables = append(variables, variable)
	}

	if err := d.Set("variable", variables); err != nil {
		return err
	}

	return nil
}

// format the strings into an id `a:b:c`
func buildId(a, b, c string) string {
	return fmt.Sprintf("%s:%s:%s", a, b, c)
}

// break string `a:b:c` into three strings `a`, `b` and `c`
func expandId(id string) (string, string, string) {
	parts := strings.SplitN(id, ":", 3)
	return parts[0], parts[1], parts[2]
}

func variableHash(v interface{}) int {
	m := v.(map[string]interface{})

	name := m["name"].(string)
	value := m["value"].(string)

	if !strings.HasPrefix(value, "xxxx") {
		value = maskCircleCiSecret(value)
	}

	return hashcode.String(
		fmt.Sprintf("%s:%s", name, value),
	)
}