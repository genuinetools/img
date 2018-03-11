package clair

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetLayer displays a Layer and optionally all of its features and vulnerabilities.
func (c *Clair) GetLayer(name string, features, vulnerabilities bool) (*Layer, error) {
	url := c.url("/v1/layers/%s?features=%t&vulnerabilities=%t", name, features, vulnerabilities)
	c.Logf("clair.layers.get url=%s name=%s", url, name)

	var respLayer layerEnvelope
	if _, err := c.getJSON(url, &respLayer); err != nil {
		return nil, err
	}

	if respLayer.Error != nil {
		return nil, fmt.Errorf("clair error: %s", respLayer.Error.Message)
	}

	return respLayer.Layer, nil
}

// PostLayer performs the analysis of a Layer from the provided path.
func (c *Clair) PostLayer(layer *Layer) (*Layer, error) {
	url := c.url("/v1/layers")
	c.Logf("clair.layers.post url=%s name=%s", url, layer.Name)

	b, err := json.Marshal(layerEnvelope{Layer: layer})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.Logf("clair.clair resp.Status=%s", resp.Status)

	var respLayer layerEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&respLayer); err != nil {
		return nil, err
	}

	if respLayer.Error != nil {
		return nil, fmt.Errorf("clair error: %s", respLayer.Error.Message)
	}

	return respLayer.Layer, err
}

// DeleteLayer removes a layer reference from clair.
func (c *Clair) DeleteLayer(name string) error {
	url := c.url("/v1/layers/%s", name)
	c.Logf("clair.layers.delete url=%s name=%s", url, name)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.Logf("clair.clair resp.Status=%s", resp.Status)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return fmt.Errorf("Got status code: %d", resp.StatusCode)
}
