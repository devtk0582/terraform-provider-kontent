package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ktangsfs/terraform-provider-kontent/client"
)

func validateCodeName(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("expected codename to be string"))
		return warns, errs
	}
	whiteSpace := regexp.MustCompile(`\s+`)
	if whiteSpace.Match([]byte(value)) {
		errs = append(errs, fmt.Errorf("codename cannot contain whitespace. Got %s", value))
		return warns, errs
	}
	return warns, errs
}

func resourceContentType() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"codename": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The codename of the content type, also acts as it's unique ID",
				ForceNew:     true,
				ValidateFunc: validateCodeName,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the content type",
			},
			"elements": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The list of elements for the content type",
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
			},
		},
		Create: resourceCreateContentType,
		Read:   resourceReadContentType,
		Update: resourceUpdateContentType,
		Delete: resourceDeleteContentType,
		Exists: resourceExistsContentType,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceCreateContentType(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	tfElements := d.Get("elements").([]interface{})
	elements := []map[string]interface{}{}

	for _, tfElement := range tfElements {
		elem := tfElement.(map[string]interface{})
		elemDto := map[string]interface{}{}
		for k, v := range elem {
			elemDto[k] = fmt.Sprintf("%v", v)
		}

		elements = append(elements, elemDto)
	}

	contentType := client.ContentType{
		Name:     d.Get("name").(string),
		CodeName: d.Get("codename").(string),
		Elements: elements,
	}

	id, err := apiClient.NewContentType(&contentType)

	if err != nil {
		return err
	}

	d.SetId(id)
	return nil
}

func resourceReadContentType(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	contentTypeId := d.Id()
	contentType, err := apiClient.GetContentType(contentTypeId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.SetId("")
		} else {
			return fmt.Errorf("error finding content type with ID %s", contentType)
		}
	}

	d.SetId(contentType.Id)
	d.Set("name", contentType.Name)
	d.Set("codename", contentType.CodeName)
	d.Set("elements", contentType.Elements)
	return nil
}

func resourceUpdateContentType(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	contentTypeId := d.Id()

	tfElements := d.Get("elements").([]interface{})
	elements := []map[string]interface{}{}

	for _, tfElement := range tfElements {
		elem := tfElement.(map[string]interface{})
		elemDto := map[string]interface{}{}
		for k, v := range elem {
			elemDto[k] = fmt.Sprintf("%v", v)
		}

		elements = append(elements, elemDto)
	}

	contentType := client.ContentType{
		Id:       contentTypeId,
		Name:     d.Get("name").(string),
		CodeName: d.Get("codename").(string),
		Elements: elements,
	}

	err := apiClient.UpdateContentType(&contentType)
	if err != nil {
		return err
	}
	d.SetId(contentType.Id)
	return nil
}

func resourceDeleteContentType(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	contentTypeId := d.Id()

	err := apiClient.DeleteContentType(contentTypeId)
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceExistsContentType(d *schema.ResourceData, m interface{}) (bool, error) {
	apiClient := m.(*client.Client)

	contentTypeId := d.Id()
	_, err := apiClient.GetContentType(contentTypeId)
	if err != nil {
		return false, err
	}
	return true, nil
}
