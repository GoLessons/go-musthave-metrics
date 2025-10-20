package test

import (
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterWithGzip(t *testing.T) {
	I := NewTester(t, nil)
	defer I.Shutdown()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := "test_counter_metric"
	value1, value2 := int64(rnd.Int31()), int64(rnd.Int31())

	t.Run("counter_not_found", func(t *testing.T) {
		getMetric := model.Metrics{
			ID:    id,
			MType: "counter",
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{
			"Accept-Encoding": "gzip",
			"Content-Type":    "application/json",
		})

		dumpErr := assert.NoError(t, err, "Ошибка при попытке сделать запрос с получением значения counter")
		require.NotNil(t, resp)
		defer resp.Body.Close()

		var result model.Metrics
		var value0 int64

		switch resp.StatusCode {
		case http.StatusOK:
			dumpErr = dumpErr && assert.Equalf(t, http.StatusOK, resp.StatusCode,
				"Несоответствие статус кода ответа ожидаемому в хендлере %q: %q", http.MethodPost, "/value/")
			dumpErr = dumpErr && assert.Containsf(t, resp.Header.Get("Content-Type"), "application/json",
				"Заголовок ответа Content-Type содержит несоответствующее значение")

			body, err := I.ReadGzip(resp)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(body, &result))

			dumpErr = dumpErr && assert.NotNil(t, result.Delta,
				"Получено не инициализированное значение Delta '%q %s'", http.MethodPost, "/value/")
			if result.Delta != nil {
				value0 = *result.Delta
			}

		case http.StatusNotFound:
			assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Ожидаем 404 для несуществующей метрики")

		default:
			dumpErr = false
			t.Fatalf("Несоответствие статус кода %d ответа ожидаемому http.StatusNotFound или http.StatusOK в хендлере %q: %q",
				resp.StatusCode, http.MethodPost, "/value/")
		}

		if !dumpErr {
			t.FailNow()
		}

		t.Logf("Первоначальное значение counter: %d", value0)
	})

	t.Run("counter_exists", func(t *testing.T) {
		c := serverModel.NewCounter(id)
		c.Inc(value1)
		require.NoError(t, I.HaveCouner(*c))

		getMetric := model.Metrics{
			ID:    id,
			MType: "counter",
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{
			"Accept-Encoding": "gzip",
			"Content-Type":    "application/json",
		})

		dumpErr := assert.NoError(t, err, "Ошибка при попытке сделать запрос с получением значения counter")
		require.NotNil(t, resp)
		defer resp.Body.Close()

		var result model.Metrics
		var retrievedValue int64

		switch resp.StatusCode {
		case http.StatusOK:
			dumpErr = dumpErr && assert.Equalf(t, http.StatusOK, resp.StatusCode,
				"Несоответствие статус кода ответа ожидаемому в хендлере %q: %q", http.MethodPost, "/value/")
			dumpErr = dumpErr && assert.Containsf(t, resp.Header.Get("Content-Type"), "application/json",
				"Заголовок ответа Content-Type содержит несоответствующее значение")
			dumpErr = dumpErr && assert.Containsf(t, resp.Header.Get("Content-Encoding"), "gzip",
				"Ответ должен быть сжат gzip")

			body, err := I.ReadGzip(resp)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(body, &result))

			dumpErr = dumpErr && assert.NotNil(t, result.Delta,
				"Получено не инициализированное значение Delta '%q %s'", http.MethodPost, "/value/")
			if result.Delta != nil {
				retrievedValue = *result.Delta
			}

			dumpErr = dumpErr && assert.Equal(t, value1, retrievedValue,
				"Полученное значение counter не соответствует ожидаемому")

		case http.StatusNotFound:
			t.Errorf("Неожиданный 404 для существующей метрики")
			dumpErr = false

		default:
			dumpErr = false
			t.Fatalf("Несоответствие статус кода %d ответа ожидаемому http.StatusNotFound или http.StatusOK в хендлере %q: %q",
				resp.StatusCode, http.MethodPost, "/value/")
		}

		if !dumpErr {
			t.FailNow()
		}

		t.Logf("Полученное значение counter: %d (ожидалось: %d)", retrievedValue, value1)
	})

	t.Run("counter_updated", func(t *testing.T) {
		c := serverModel.NewCounter(id)
		c.Inc(value1 + value2)
		require.NoError(t, I.HaveCouner(*c))

		getMetric := model.Metrics{
			ID:    id,
			MType: "counter",
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{
			"Accept-Encoding": "gzip",
			"Content-Type":    "application/json",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Containsf(t, resp.Header.Get("Content-Type"), "application/json",
			"Content-Type должен быть application/json")
		assert.Containsf(t, resp.Header.Get("Content-Encoding"), "gzip",
			"Ответ должен быть сжат gzip")

		body, err := I.ReadGzip(resp)
		require.NoError(t, err)

		var result model.Metrics
		require.NoError(t, json.Unmarshal(body, &result))

		require.NotNil(t, result.Delta, "Delta не должна быть nil")
		assert.Equal(t, value1+value2, *result.Delta, "Значение должно соответствовать обновленному")

		t.Logf("Обновленное значение counter: %d (ожидалось: %d)", *result.Delta, value1+value2)
	})
}
