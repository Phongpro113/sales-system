package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func getProduct(id uint) (*Product, error) {
	url := fmt.Sprintf("%s/api/products/%d", productServiceURL, id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("product service returned %d: %s", resp.StatusCode, string(body))
	}

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode product: %w", err)
	}

	return &product, nil
}

func updateProductStock(productID uint, quantity int) error {
	url := fmt.Sprintf("%s/api/products/%d/stock", productServiceURL, productID)

	reqBody := map[string]int{"quantity": -quantity}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stock update failed: %s", string(body))
	}

	return nil
}

func enrichItemsWithProductNames(items []OrderItem) []OrderItem {
	for i := range items {
		if items[i].ProductName != "" {
			items[i].Subtotal = items[i].Price * float64(items[i].Quantity)
			continue
		}
		product, err := getProduct(items[i].ProductID)
		if err == nil {
			items[i].ProductName = product.Name
		} else {
			items[i].ProductName = fmt.Sprintf("Product #%d", items[i].ProductID)
		}
		items[i].Subtotal = items[i].Price * float64(items[i].Quantity)
	}
	return items
}
