#!/bin/env node

var fs = require('fs');

function fixed(num) {
  return parseFloat(num.toFixed(2));
}

// inital setup
freeMemory = {
  id: "free_memory",
  data: []
};

cpuUsage = {
  id: "cpu_usage",
  data: []
};

// number of data points
var N = 100;

// start time [ now - num of data points ]
var t = new Date().getTime() - N * 30 * 1000;

// fill random data
for (i = 0; i < N; i+=1) {
  freeMemory.data.push({
    timestamp: t + i * 30 * 1000,
    value: fixed(1500000 + Math.random() * 500000)
  });

  cpuUsage.data.push({
    timestamp: t + i * 30 * 1000,
    value: fixed(1000 + Math.random() * 500)
  });
}

// print out
fs.writeFile('test-data.json', JSON.stringify([freeMemory, cpuUsage], null, 2));
