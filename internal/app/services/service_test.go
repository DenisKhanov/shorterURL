package services

import (
	"context"
	"errors"
	"github.com/DenisKhanov/shorterURL/internal/app/services/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRepository(ctrl)
	mockEncoder := mocks.NewMockEncoder(ctrl)
	baseURL := "http://localhost:8080"
	service := NewShortURLServices(mockRepo, mockEncoder, baseURL)
	if service.repository != mockRepo {
		t.Errorf("Expected repository to be set, got %v", service.repository)
	}
	if service.encoder != mockEncoder {
		t.Errorf("Expected encoder to be set, got %v", service.encoder)
	}
}

func TestServices_GetShortURL(t *testing.T) {

	tests := []struct {
		name             string
		originalURL      string
		expectedShortURL string
		mockSetup        func(mockRepo *mocks.MockRepository, mockEncoder *mocks.MockEncoder)
	}{
		{
			name:             "ShortURL found in repository",
			originalURL:      "http://original.url",
			expectedShortURL: "http://localhost:8080/shortURL",
			mockSetup: func(mockRepo *mocks.MockRepository, mockEncoder *mocks.MockEncoder) {
				mockRepo.EXPECT().GetShortURLFromDB(gomock.Any(), "http://original.url").Return("shortURL", nil).AnyTimes()
			},
		},
		{
			name:             "ShortURL not found in repository",
			originalURL:      "http://original.url",
			expectedShortURL: "http://localhost:8080/shortURL",
			mockSetup: func(mockRepo *mocks.MockRepository, mockEncoder *mocks.MockEncoder) {
				mockRepo.EXPECT().GetShortURLFromDB(context.Background(), "http://original.url").Return("", errors.New("short URL not found")).AnyTimes()
				mockEncoder.EXPECT().CryptoBase62Encode().Return("shortURL").AnyTimes()
				mockRepo.EXPECT().StoreURLInDB(gomock.Any(), "http://original.url", "shortURL").Return(nil).AnyTimes()
			},
		},
		{
			name: "ShortURL not found in repository " +
				"and failed to save",
			originalURL:      "http://original.url",
			expectedShortURL: "",
			mockSetup: func(mockRepo *mocks.MockRepository, mockEncoder *mocks.MockEncoder) {
				mockRepo.EXPECT().GetShortURLFromDB(context.Background(), "http://original.url").Return("", errors.New("short URL not found")).AnyTimes()
				mockEncoder.EXPECT().CryptoBase62Encode().Return("shortURL").AnyTimes()
				mockRepo.EXPECT().StoreURLInDB(gomock.Any(), "http://original.url", "shortURL").Return(errors.New("error saving shortUrl")).AnyTimes()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mocks.NewMockRepository(ctrl)
			mockEncoder := mocks.NewMockEncoder(ctrl)
			tt.mockSetup(mockRepo, mockEncoder)
			service := ShortURLServices{repository: mockRepo, encoder: mockEncoder, baseURL: "http://localhost:8080"}
			result, err := service.GetShortURL(context.Background(), tt.originalURL)
			if tt.name == "ShortURL found in repository" {
				assert.Equal(t, tt.expectedShortURL, result)
				assert.EqualError(t, err, "short URL found in database")
			} else {
				if tt.name == "ShortURL not found in repository "+
					"and failed to save" {
					assert.EqualError(t, err, "error saving shortUrl")
				} else {
					assert.NoError(t, err)
				}
			}

			assert.Equal(t, tt.expectedShortURL, result)
		})
	}

}

func TestServices_GetOriginalURL(t *testing.T) {
	tests := []struct {
		name                string
		shortURL            string
		expectedOriginalURL string
		mockSetup           func(mockRepo *mocks.MockRepository)
	}{
		{
			name:                "OriginalURL found in repository",
			shortURL:            "shortURL",
			expectedOriginalURL: "http://original.url",
			mockSetup: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().GetOriginalURLFromDB(gomock.Any(), "shortURL").Return("http://original.url", nil).AnyTimes()
			},
		},
		{
			name:                "OriginalURL not found in repository",
			shortURL:            "shortURL",
			expectedOriginalURL: "",
			mockSetup: func(mockRepo *mocks.MockRepository) {
				mockRepo.EXPECT().GetOriginalURLFromDB(gomock.Any(), "shortURL").Return("", errors.New("original URL not found")).AnyTimes()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mocks.NewMockRepository(ctrl)
			tt.mockSetup(mockRepo)
			service := ShortURLServices{repository: mockRepo, baseURL: "http://localhost:8080"}
			result, err := service.GetOriginalURL(context.Background(), tt.shortURL)
			if tt.name == "OriginalURL not found in repository" {
				assert.EqualError(t, err, "original URL not found")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedOriginalURL, result)
		})
	}
}

func TestCryptoBase62Encode(t *testing.T) {
	service := ShortURLServices{}

	encoded := service.CryptoBase62Encode()

	if len(encoded) > 8 {
		t.Errorf("expected length <= 8, got %d", len(encoded))
	}

	const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for _, char := range encoded {
		if !strings.ContainsRune(base62Chars, char) {
			t.Errorf("invalid character %c in encoded string", char)
		}
	}
}

func BenchmarkShortURLServices_CryptoBase62Encode(b *testing.B) {
	service := ShortURLServices{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CryptoBase62Encode()
	}
}
