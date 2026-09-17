"use client";

export type BodyMeasurement = {
  id: string;
  height_cm: number;
  weight_kg: number;
  bmi: number;
  age_years: number;
  bmi_percentile?: number;
  healthy_bmi_min?: number;
  healthy_bmi_max?: number;
  reference: string;
  status: string;
  measured_at: string;
};

export const bmiStatus: Record<string, { label: string; color: string }> = {
  underweight: { label: "کمبود وزن", color: "#3b82f6" },
  healthy: { label: "محدوده سالم", color: "#16a36a" },
  overweight: { label: "اضافه‌وزن", color: "#e5a51a" },
  obesity_1: { label: "چاقی درجه یک", color: "#ef7d32" },
  obesity_2: { label: "چاقی درجه دو", color: "#e65353" },
  obesity_3: { label: "چاقی درجه سه", color: "#9f3c74" },
  severe_obesity: { label: "چاقی شدید", color: "#9f3c74" },
  age_specific: { label: "نیازمند ارزیابی سنی", color: "#7767c7" },
};

export default function BMITrendChart({
  measurements,
  compact = false,
}: {
  measurements: BodyMeasurement[];
  compact?: boolean;
}) {
  const points = [...measurements].reverse().slice(compact ? -10 : -24);
  if (!points.length) {
    return (
      <div className="bmi-chart-empty">
        <span>نمودار پس از اولین ثبت ساخته می‌شود.</span>
      </div>
    );
  }
  const healthyValues = points.flatMap((item) =>
    item.healthy_bmi_min && item.healthy_bmi_max ? [item.healthy_bmi_min, item.healthy_bmi_max] : []
  );
  const values = [...points.map((item) => item.bmi), ...healthyValues];
  const min = Math.max(10, Math.floor(Math.min(...values) - 2));
  const max = Math.min(55, Math.ceil(Math.max(...values) + 2));
  const range = Math.max(max - min, 4);
  const width = 720;
  const height = compact ? 180 : 250;
  const paddingX = 34;
  const paddingY = 28;
  const x = (index: number) =>
    points.length === 1
      ? width / 2
      : paddingX + (index * (width - paddingX * 2)) / (points.length - 1);
  const y = (value: number) => paddingY + ((max - value) * (height - paddingY * 2)) / range;
  const path = points.map((item, index) => `${x(index)},${y(item.bmi)}`).join(" ");
  const area = `${paddingX},${height - paddingY} ${path} ${width - paddingX},${height - paddingY}`;
  const allHaveHealthyRange = points.every((item) => item.healthy_bmi_min && item.healthy_bmi_max);
  const healthyArea = allHaveHealthyRange
    ? [
        ...points.map((item, index) => `${x(index)},${y(item.healthy_bmi_max!)}`),
        ...[...points].reverse().map((item, reverseIndex) => {
          const index = points.length - 1 - reverseIndex;
          return `${x(index)},${y(item.healthy_bmi_min!)}`;
        }),
      ].join(" ")
    : "";
  return (
    <div className={`bmi-chart${compact ? " compact" : ""}`}>
      <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="نمودار تغییرات شاخص توده بدنی">
        {[0, 1, 2, 3].map((line) => {
          const lineY = paddingY + (line * (height - paddingY * 2)) / 3;
          return <line key={line} x1={paddingX} x2={width - paddingX} y1={lineY} y2={lineY} />;
        })}
        {healthyArea && <polygon className="bmi-healthy-zone" points={healthyArea} />}
        <defs>
          <linearGradient id={compact ? "bmiAreaCompact" : "bmiArea"} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stopColor="#5b8def" stopOpacity=".3" />
            <stop offset="1" stopColor="#5b8def" stopOpacity="0" />
          </linearGradient>
        </defs>
        <polygon points={area} fill={`url(#${compact ? "bmiAreaCompact" : "bmiArea"})`} />
        <polyline className="bmi-chart-line" points={path} />
        {points.map((item, index) => (
          <g key={item.id}>
            <circle cx={x(index)} cy={y(item.bmi)} r={compact ? 5 : 6} />
            {!compact && (
              <text x={x(index)} y={y(item.bmi) - 13}>
                {item.bmi.toLocaleString("fa-IR")}
              </text>
            )}
          </g>
        ))}
      </svg>
      {healthyArea && <span className="bmi-healthy-zone-label">محدوده سالم متناسب با سن</span>}
      <div className="bmi-chart-labels">
        <span>{new Date(points[0].measured_at).toLocaleDateString("fa-IR-u-ca-persian")}</span>
        <span>{new Date(points.at(-1)!.measured_at).toLocaleDateString("fa-IR-u-ca-persian")}</span>
      </div>
    </div>
  );
}
