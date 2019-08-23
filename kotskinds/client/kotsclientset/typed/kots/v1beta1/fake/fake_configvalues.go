/*
Copyright 2019 Replicated, Inc..

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeConfigValueses implements ConfigValuesInterface
type FakeConfigValueses struct {
	Fake *FakeKotsV1beta1
	ns   string
}

var configvaluesesResource = schema.GroupVersionResource{Group: "kots.io", Version: "v1beta1", Resource: "configvalueses"}

var configvaluesesKind = schema.GroupVersionKind{Group: "kots.io", Version: "v1beta1", Kind: "ConfigValues"}

// Get takes name of the configValues, and returns the corresponding configValues object, and an error if there is any.
func (c *FakeConfigValueses) Get(name string, options v1.GetOptions) (result *v1beta1.ConfigValues, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(configvaluesesResource, c.ns, name), &v1beta1.ConfigValues{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ConfigValues), err
}

// List takes label and field selectors, and returns the list of ConfigValueses that match those selectors.
func (c *FakeConfigValueses) List(opts v1.ListOptions) (result *v1beta1.ConfigValuesList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(configvaluesesResource, configvaluesesKind, c.ns, opts), &v1beta1.ConfigValuesList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ConfigValuesList{ListMeta: obj.(*v1beta1.ConfigValuesList).ListMeta}
	for _, item := range obj.(*v1beta1.ConfigValuesList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested configValueses.
func (c *FakeConfigValueses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(configvaluesesResource, c.ns, opts))

}

// Create takes the representation of a configValues and creates it.  Returns the server's representation of the configValues, and an error, if there is any.
func (c *FakeConfigValueses) Create(configValues *v1beta1.ConfigValues) (result *v1beta1.ConfigValues, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(configvaluesesResource, c.ns, configValues), &v1beta1.ConfigValues{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ConfigValues), err
}

// Update takes the representation of a configValues and updates it. Returns the server's representation of the configValues, and an error, if there is any.
func (c *FakeConfigValueses) Update(configValues *v1beta1.ConfigValues) (result *v1beta1.ConfigValues, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(configvaluesesResource, c.ns, configValues), &v1beta1.ConfigValues{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ConfigValues), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeConfigValueses) UpdateStatus(configValues *v1beta1.ConfigValues) (*v1beta1.ConfigValues, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(configvaluesesResource, "status", c.ns, configValues), &v1beta1.ConfigValues{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ConfigValues), err
}

// Delete takes name of the configValues and deletes it. Returns an error if one occurs.
func (c *FakeConfigValueses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(configvaluesesResource, c.ns, name), &v1beta1.ConfigValues{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConfigValueses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(configvaluesesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ConfigValuesList{})
	return err
}

// Patch applies the patch and returns the patched configValues.
func (c *FakeConfigValueses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ConfigValues, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(configvaluesesResource, c.ns, name, pt, data, subresources...), &v1beta1.ConfigValues{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ConfigValues), err
}
