  <template>
    <div class="chart-wrapper">
      <div class="chart-section">
        <div class="radial-container">
          <apexchart
            type="radialBar"
            height="350"
            :options="radialOptions"
            :series="radialSeries"
          ></apexchart>
        </div>
      </div>

      <div class="legend-section">
        <div v-for="(item, index) in driveLegends" :key="index" class="legend-item">
          <span class="legend-dot" :style="{ backgroundColor: item.color }"></span>
          <div class="legend-info">
            <span class="legend-name">{{ item.name }}</span>
            <span class="legend-size">{{ item.used }} / {{ item.total }}</span>
          </div>
        </div>
      </div>
    </div>
  </template>

  <script setup>
  import { computed } from "vue";
  import VueApexCharts from "vue3-apexcharts";

  const apexchart = VueApexCharts;

  const props = defineProps({
    series: { type: Array, required: true },
    labels: { type: Array, required: true },
    colors: { type: Array, required: true }
  });

  // Hitung total used dan free dari semua drive
  const totalUsed = computed(() => {
    let used = 0;
    props.labels.forEach((label, index) => {
      if (!label.includes("(Free)")) {
        used += props.series[index];
      }
    });
    return used;
  });

  const totalCapacity = computed(() => {
    return props.series.reduce((a, b) => a + b, 0);
  });

  const totalFree = computed(() => {
    return totalCapacity.value - totalUsed.value;
  });

  // Radial bar series dengan percentage
  const radialSeries = computed(() => {
    const series = [];
    
    // Add setiap drive
    props.labels.forEach((label, index) => {
      if (!label.includes("(Free)")) {
        const used = props.series[index];
        const freeIndex = props.labels.indexOf(label + " (Free)");
        const free = freeIndex !== -1 ? props.series[freeIndex] : 0;
        const total = used + free;
        const percentage = total > 0 ? Math.round((used / total) * 100) : 0;
        series.push(percentage);
      }
    });
    
    // Add free space dari total capacity
    if (totalCapacity.value > 0) {
      const freePercentage = Math.round((totalFree.value / totalCapacity.value) * 100);
      series.push(freePercentage);
    }
    
    return series;
  });

  // Radial labels
  const radialLabels = computed(() => {
    const labels = [];
    
    props.labels.forEach((label, index) => {
      if (!label.includes("(Free)")) {
        labels.push(label);
      }
    });
    
    labels.push("Free Space");
    return labels;
  });

  // Radial colors
  const radialColors = computed(() => {
    const colors = [];
    
    props.labels.forEach((label, index) => {
      if (!label.includes("(Free)")) {
        colors.push(props.colors[index]);
      }
    });
    
    colors.push("#d3d3d3");
    return colors;
  });

  // Legend items untuk drive
  const driveLegends = computed(() => {
    const legends = [];
    props.labels.forEach((label, index) => {
      if (!label.includes("(Free)")) {
        const used = props.series[index];
        const freeIndex = props.labels.indexOf(label + " (Free)");
        const free = freeIndex !== -1 ? props.series[freeIndex] : 0;
        
        legends.push({
          name: label,
          used: formatSize(used),
          total: formatSize(used + free),
          color: props.colors[index]
        });
      }
    });
    
    // Add free space
    legends.push({
      name: "Free Space",
      used: formatSize(totalFree.value),
      total: formatSize(totalCapacity.value),
      color: "#d3d3d3"
    });
    
    return legends;
  });

  function formatSize(bytes) {
    if (bytes >= 1024) {
      return (bytes / 1024).toFixed(2) + " TB";
    }
    return bytes.toFixed(2) + " GB";
  }

  const radialOptions = computed(() => ({
    chart: {
      type: "radialBar",
      foreColor: "#1a1a1a",
      toolbar: { show: false },
      animations: {
        enabled: true,
        speed: 800,
        animateGradually: {
          enabled: true,
          delay: 150
        }
      }
    },

    labels: radialLabels.value,
    colors: radialColors.value,

    plotOptions: {
      radialBar: {
        size: undefined,
        inverseOrder: false,
        hollow: {
          margin: 5,
          size: "30%",
          background: "transparent",
          image: undefined
        },
        track: {
          show: true,
          background: "#f5f5f5",
          strokeWidth: "95%",
          opacity: 1,
          margin: 8
        },
        dataLabels: {
          name: {
            show: true,
            fontSize: "12px",
            fontWeight: 500,
            color: "#1a1a1a",
            offsetY: -10
          },
          value: {
            show: true,
            fontSize: "14px",
            fontWeight: "bold",
            color: "#1a1a1a",
            offsetY: 5,
            formatter: function(val) {
              return Math.round(val) + "%";
            }
          }
        }
      }
    },

    states: {
      hover: {
        filter: { type: "none" }
      },
      active: {
        filter: { type: "none" }
      }
    },

    stroke: {
      lineCap: "round"
    },

    legend: {
      show: false
    },

    tooltip: {
      enabled: true,
      theme: "light",
      y: {
        formatter: function(val) {
          return Math.round(val) + "%";
        }
      }
    }
  }));
  </script>

  <style scoped>
  .chart-wrapper {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 2rem;
    padding: 0;
  }

  .chart-section {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .radial-container {
    width: 100%;
    height: 350px;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  /* Legend Section */
  .legend-section {
    flex: 0 0 auto;
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    max-height: 350px;
    overflow-y: auto;
    padding-right: 0.5rem;
  }

  .legend-item {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.5rem;
    border-radius: 4px;
    transition: background-color 0.2s ease;
  }

  .legend-item:hover {
    background-color: #f5f5f5;
  }

  .legend-dot {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    flex-shrink: 0;
    margin-top: 0.25rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .legend-info {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
  }

  .legend-name {
    font-size: 0.85rem;
    font-weight: 600;
    color: #1a1a1a;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .legend-size {
    font-size: 0.75rem;
    color: #999;
    font-weight: 400;
    font-family: monospace;
  }

  /* Scrollbar */
  .legend-section::-webkit-scrollbar {
    width: 4px;
  }

  .legend-section::-webkit-scrollbar-track {
    background: transparent;
  }

  .legend-section::-webkit-scrollbar-thumb {
    background: #d0d0d0;
    border-radius: 2px;
  }

  .legend-section::-webkit-scrollbar-thumb:hover {
    background: #999;
  }

  /* Responsive */
  @media (max-width: 1024px) {
    .chart-wrapper {
      flex-direction: column;
      gap: 1.5rem;
    }

    .radial-container {
      height: 300px;
    }

    .legend-section {
      max-height: 200px;
      flex-direction: row;
      flex-wrap: wrap;
    }

    .legend-item {
      flex: 0 0 calc(50% - 0.4rem);
    }
  }

  @media (max-width: 768px) {
    .chart-wrapper {
      gap: 1rem;
    }

    .radial-container {
      height: 250px;
    }

    .legend-section {
      max-height: 150px;
      gap: 0.5rem;
    }

    .legend-item {
      flex: 0 0 auto;
    }
  }
  </style>