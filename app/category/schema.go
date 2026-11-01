package category

import "time"

type CategoryRead struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	CategoryType CategoryType `json:"type"`
	OwnerID      string       `json:"ownerID"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

func BuildCategoryReadList(categorys []Category) []CategoryRead {
	categoryResponse := make([]CategoryRead, 0, len(categorys))
	for _, category := range categorys {
		schemaCategory := BuildCategoryRead(category)
		categoryResponse = append(categoryResponse, schemaCategory)
	}
	return categoryResponse
}

func BuildCategoryRead(category Category) CategoryRead {
	return CategoryRead{
		ID:           category.ID,
		Name:         category.Name,
		CategoryType: category.Type,
		OwnerID:      category.OwnerID,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
	}
}
