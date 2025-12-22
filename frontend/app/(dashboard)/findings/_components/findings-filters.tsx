"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MagnifyingGlassIcon, XIcon } from "@phosphor-icons/react";

const severityOptions = [
  { value: "all", label: "All Severities" },
  { value: "CRITICAL", label: "Critical" },
  { value: "HIGH", label: "High" },
  { value: "MEDIUM", label: "Medium" },
  { value: "LOW", label: "Low" },
];

const serviceOptions = [
  { value: "all", label: "All Services" },
  { value: "S3", label: "S3" },
  { value: "EC2", label: "EC2" },
  { value: "IAM", label: "IAM" },
  { value: "Lambda", label: "Lambda" },
  { value: "ECS", label: "ECS" },
  { value: "DynamoDB", label: "DynamoDB" },
];

const complianceOptions = [
  { value: "all", label: "All Frameworks" },
  { value: "CIS", label: "CIS Benchmark" },
  { value: "SOC2", label: "SOC 2" },
  { value: "PCI-DSS", label: "PCI-DSS" },
  { value: "GDPR", label: "GDPR" },
  { value: "NIST", label: "NIST 800-53" },
];

export function FindingsFilters() {
  const [search, setSearch] = React.useState("");
  const [severity, setSeverity] = React.useState("all");
  const [service, setService] = React.useState("all");
  const [compliance, setCompliance] = React.useState("all");

  const activeFilters = [
    severity !== "all" && { key: "severity", value: severity },
    service !== "all" && { key: "service", value: service },
    compliance !== "all" && { key: "compliance", value: compliance },
  ].filter(Boolean) as Array<{ key: string; value: string }>;

  const clearFilter = (key: string) => {
    switch (key) {
      case "severity":
        setSeverity("all");
        break;
      case "service":
        setService("all");
        break;
      case "compliance":
        setCompliance("all");
        break;
    }
  };

  const clearAllFilters = () => {
    setSearch("");
    setSeverity("all");
    setService("all");
    setCompliance("all");
  };

  return (
    <div className="space-y-4 mb-6">
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <Input
            placeholder="Search findings..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex gap-2">
          <Select value={severity} onValueChange={setSeverity}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="Severity" />
            </SelectTrigger>
            <SelectContent>
              {severityOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={service} onValueChange={setService}>
            <SelectTrigger className="w-[130px]">
              <SelectValue placeholder="Service" />
            </SelectTrigger>
            <SelectContent>
              {serviceOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={compliance} onValueChange={setCompliance}>
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Compliance" />
            </SelectTrigger>
            <SelectContent>
              {complianceOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {activeFilters.length > 0 && (
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Active filters:</span>
          {activeFilters.map((filter) => (
            <Badge key={filter.key} variant="secondary" className="gap-1">
              {filter.value}
              <button
                onClick={() => clearFilter(filter.key)}
                className="ml-1 hover:text-foreground"
              >
                <XIcon className="size-3" />
              </button>
            </Badge>
          ))}
          <Button
            variant="ghost"
            size="sm"
            onClick={clearAllFilters}
            className="text-muted-foreground"
          >
            Clear all
          </Button>
        </div>
      )}
    </div>
  );
}
