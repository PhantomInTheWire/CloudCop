"use client";

import { Bar, BarChart, XAxis, YAxis, Cell } from "recharts";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";

const chartConfig = {
  count: {
    label: "Findings",
  },
  critical: {
    label: "Critical",
    color: "hsl(0, 84%, 60%)",
  },
  high: {
    label: "High",
    color: "hsl(25, 95%, 53%)",
  },
  medium: {
    label: "Medium",
    color: "hsl(45, 93%, 47%)",
  },
  low: {
    label: "Low",
    color: "hsl(142, 76%, 36%)",
  },
} satisfies ChartConfig;

interface FindingsBySeverityChartProps {
  critical: number;
  high: number;
  medium: number;
  low: number;
}

export function FindingsBySeverityChart({
  critical,
  high,
  medium,
  low,
}: FindingsBySeverityChartProps) {
  const chartData = [
    { severity: "Critical", count: critical, fill: chartConfig.critical.color },
    { severity: "High", count: high, fill: chartConfig.high.color },
    { severity: "Medium", count: medium, fill: chartConfig.medium.color },
    { severity: "Low", count: low, fill: chartConfig.low.color },
  ];

  return (
    <ChartContainer config={chartConfig} className="h-[200px] w-full">
      <BarChart
        data={chartData}
        layout="vertical"
        margin={{ left: 0, right: 20 }}
      >
        <YAxis
          dataKey="severity"
          type="category"
          tickLine={false}
          axisLine={false}
          width={70}
          tick={{ fontSize: 12 }}
        />
        <XAxis type="number" hide />
        <ChartTooltip
          cursor={false}
          content={<ChartTooltipContent hideLabel />}
        />
        <Bar dataKey="count" radius={[0, 4, 4, 0]}>
          {chartData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
        </Bar>
      </BarChart>
    </ChartContainer>
  );
}
