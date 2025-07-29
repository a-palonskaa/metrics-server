import http from 'k6/http';
import { hmac } from 'k6/crypto';
import pako from 'https://cdn.skypack.dev/pako';
import { randomIntBetween, randomString  } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

function gzipSync(data) {
  return pako.gzip(data);
}

export default function () {
  const metricTypes = ['gauge', 'counter'];

  const randomMetric1 = metricTypes[randomIntBetween(0, metricTypes.length - 1)];
  const randomValue1 = randomIntBetween(1, 1000);
  const randomID1 = randomString(5);

  const randomMetric2 = metricTypes[randomIntBetween(0, metricTypes.length - 1)];
  const randomValue2 = randomIntBetween(1, 1000);
  const randomID2 = randomString(5);

  const randomID3 = randomString(7);
  const randomVal3 = randomIntBetween(1, 1000);

  const randomID4 = randomString(7);
  const randomVal4 = randomIntBetween(1, 1000);

  const randomMetric = randomString(10);

  const randomPath = randomString(randomIntBetween(1, 50));

  const createMetricBody = (id, type, value) => {
  const body = {
    id: id,
    type: type
  };

  if (type === 'gauge') {
    const val = parseFloat(value);
    if (!isNaN(val)) {
      body.value = val;
    }
  } else if (type === 'counter') {
    const delta = parseInt(value);
    if (!isNaN(delta)) {
      body.delta = delta;
    }
  }

  return body;
};

const prepareRequestBody = (body) => {
  try {
    const jsonBody = JSON.stringify(body);
    const signature = hmac('sha256', "aliffka", jsonBody, 'hex');
    const compressedBody = gzipSync(jsonBody);

    return {
      body: compressedBody,
      headers: {
        'Content-Type': 'application/json',
        'Content-Encoding': 'gzip',
        'Accept-Encoding': 'gzip',
        'HashSHA256': signature
      }
    };
  } catch (e) {
    console.error("Error in prepareRequestBody:", e && e.stack ? e.stack : e);
    return { body: JSON.stringify(body), headers: { 'Content-Type': 'application/json' } };
  }
};

  const endpoints = [
    {
      url: `http://127.0.0.1:8080/value/${randomMetric1}/${randomID1}`,
      method: 'GET'
    },
    {
      url: 'http://127.0.0.1:8080/update/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: 'http://127.0.0.1:8080/update/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: `http://127.0.0.1:8080/value/`,
      method: 'POST',
      body: createMetricBody(randomID2, randomMetric2, randomValue2)
    },
    {
      url: 'http://127.0.0.1:8080/update/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: `http://127.0.0.1:8080/value/gauge/${randomPath}`,
      method: 'GET'
    },
    {
      url: 'http://127.0.0.1:8080/value/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: `http://127.0.0.1:8080/ping`,
      method: 'GET'
    },
    {
      url: 'http://127.0.0.1:8080/value/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: 'http://127.0.0.1:8080/value/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: `http://127.0.0.1:8080/value/`,
      method: 'GET'
    },
    {
      url: 'http://127.0.0.1:8080/updates/',
      method: 'POST',
      body: [
        createMetricBody(randomID1, randomMetric1, randomValue1),
        createMetricBody(randomID2, randomMetric2, randomValue2)
      ]
    },
    {
      url: 'http://127.0.0.1:8080/value/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1)
    },
    {
      url: `http://127.0.0.1:8080/`,
      method: 'GET'
    },
    {
      url: `http://127.0.0.1:8080/update/${randomMetric2}/${randomID4}/${randomVal4}`,
      method: 'POST'
    },
    {
      url: `http://127.0.0.1:8080/value/${randomMetric2}/${randomID4}`,
      method: 'GET'
    },
    {
      url: 'http://127.0.0.1:8080/value/',
      method: 'POST',
      body: createMetricBody(randomID1, randomMetric1, randomValue1),
    },
    {
      url: `http://127.0.0.1:8080/update/${randomMetric1}/${randomID3}/${randomVal3}`,
      method: 'POST'
    },
    {
      url: `http://127.0.0.1:8080/value/gauge/${randomPath}`,
      method: 'GET'
    },
    {
      url: 'http://127.0.0.1:8080/updates/',
      method: 'POST',
      body: [
        createMetricBody(randomID1, randomMetric1, randomValue1),
        createMetricBody(randomID2, randomMetric2, randomValue2)
      ]
    },
    {
      url: 'http://127.0.0.1:8080/updates/',
      method: 'POST',
      body: [
        createMetricBody(randomID3, randomMetric1, randomValue1),
        createMetricBody(randomID4, randomMetric2, randomValue2)
      ]
    },
    {
      url: `http://127.0.0.1:8080/value/`,
      method: 'POST',
      body: createMetricBody(randomID3, randomMetric1, randomValue1),
    },
  ];

  for (let endpoint of endpoints) {
    let requestParams = {
      headers: { 'Content-Type': 'application/json' }
    };

  if (endpoint.body) {
  if (Array.isArray(endpoint.body)) {
    requestParams.body = JSON.stringify(endpoint.body);
    requestParams.headers['Content-Type'] = 'application/json';
  } else {
    const prepared = prepareRequestBody(endpoint.body);
    requestParams.body = prepared.body;

    // Сохраняем headers и дополняем вручную
    requestParams.headers = Object.assign({}, requestParams.headers, prepared.headers);
  }
  }

    const res = http.request(
      endpoint.method,
      endpoint.url,
      requestParams.body || null,
      requestParams
    );
  }
}
