"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PlayIcon, CheckIcon, CloudIcon } from "@phosphor-icons/react";

const awsServices = [
  { id: "s3", name: "S3", description: "Simple Storage Service" },
  { id: "ec2", name: "EC2", description: "Elastic Compute Cloud" },
  { id: "iam", name: "IAM", description: "Identity & Access Management" },
  { id: "lambda", name: "Lambda", description: "Serverless Functions" },
  { id: "ecs", name: "ECS", description: "Elastic Container Service" },
  { id: "dynamodb", name: "DynamoDB", description: "NoSQL Database" },
];

const awsRegions = [
  { id: "us-east-1", name: "US East (N. Virginia)" },
  { id: "us-west-2", name: "US West (Oregon)" },
  { id: "eu-west-1", name: "Europe (Ireland)" },
  { id: "eu-central-1", name: "Europe (Frankfurt)" },
  { id: "ap-southeast-1", name: "Asia Pacific (Singapore)" },
  { id: "ap-northeast-1", name: "Asia Pacific (Tokyo)" },
];

const awsAccounts = [
  { id: "123456789012", name: "Production", verified: true },
  { id: "987654321098", name: "Development", verified: true },
];

export function NewScanForm() {
  const [selectedAccount, setSelectedAccount] = React.useState<string>("");
  const [selectedServices, setSelectedServices] = React.useState<string[]>([
    "s3",
    "ec2",
    "iam",
  ]);
  const [selectedRegions, setSelectedRegions] = React.useState<string[]>([
    "us-east-1",
  ]);
  const [isScanning, setIsScanning] = React.useState(false);

  const toggleService = (serviceId: string) => {
    setSelectedServices((prev) =>
      prev.includes(serviceId)
        ? prev.filter((id) => id !== serviceId)
        : [...prev, serviceId],
    );
  };

  const toggleRegion = (regionId: string) => {
    setSelectedRegions((prev) =>
      prev.includes(regionId)
        ? prev.filter((id) => id !== regionId)
        : [...prev, regionId],
    );
  };

  const handleStartScan = async () => {
    setIsScanning(true);
    // TODO: Call GraphQL mutation to start scan
    console.log("Starting scan with:", {
      account: selectedAccount,
      services: selectedServices,
      regions: selectedRegions,
    });
    // Simulate scan start
    setTimeout(() => {
      setIsScanning(false);
    }, 2000);
  };

  return (
    <FieldGroup className="max-w-2xl">
      <Field>
        <FieldLabel>AWS Account</FieldLabel>
        <Select value={selectedAccount} onValueChange={setSelectedAccount}>
          <SelectTrigger>
            <SelectValue placeholder="Select an AWS account" />
          </SelectTrigger>
          <SelectContent>
            {awsAccounts.map((account) => (
              <SelectItem key={account.id} value={account.id}>
                <div className="flex items-center gap-2">
                  <CloudIcon className="size-4" />
                  <span>{account.name}</span>
                  <span className="text-muted-foreground">({account.id})</span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </Field>

      <Field>
        <FieldLabel>Services to Scan</FieldLabel>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 mt-2">
          {awsServices.map((service) => {
            const isSelected = selectedServices.includes(service.id);
            return (
              <button
                key={service.id}
                type="button"
                onClick={() => toggleService(service.id)}
                className={`flex items-center justify-between gap-2 rounded-lg border p-3 text-left transition-colors ${
                  isSelected
                    ? "border-primary bg-primary/5"
                    : "border-border hover:bg-muted/50"
                }`}
              >
                <div>
                  <div className="font-medium text-sm">{service.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {service.description}
                  </div>
                </div>
                {isSelected && (
                  <CheckIcon className="size-4 text-primary" weight="bold" />
                )}
              </button>
            );
          })}
        </div>
        <div className="mt-2 flex gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setSelectedServices(awsServices.map((s) => s.id))}
          >
            Select All
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setSelectedServices([])}
          >
            Clear All
          </Button>
        </div>
      </Field>

      <Field>
        <FieldLabel>AWS Regions</FieldLabel>
        <div className="flex flex-wrap gap-2 mt-2">
          {awsRegions.map((region) => {
            const isSelected = selectedRegions.includes(region.id);
            return (
              <Badge
                key={region.id}
                variant={isSelected ? "default" : "outline"}
                className="cursor-pointer"
                onClick={() => toggleRegion(region.id)}
              >
                {isSelected && (
                  <CheckIcon className="mr-1 size-3" weight="bold" />
                )}
                {region.name}
              </Badge>
            );
          })}
        </div>
        <div className="mt-2 flex gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setSelectedRegions(awsRegions.map((r) => r.id))}
          >
            Select All
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setSelectedRegions([])}
          >
            Clear All
          </Button>
        </div>
      </Field>

      <div className="pt-4 border-t">
        <Button
          size="lg"
          disabled={
            !selectedAccount ||
            selectedServices.length === 0 ||
            selectedRegions.length === 0 ||
            isScanning
          }
          onClick={handleStartScan}
        >
          {isScanning ? (
            <>
              <span className="animate-spin mr-2">
                <PlayIcon className="size-5" />
              </span>
              Starting Scan...
            </>
          ) : (
            <>
              <PlayIcon className="mr-2 size-5" />
              Start Security Scan
            </>
          )}
        </Button>
        <p className="text-sm text-muted-foreground mt-2">
          Estimated duration: 3-5 minutes for {selectedServices.length}{" "}
          service(s) across {selectedRegions.length} region(s)
        </p>
      </div>
    </FieldGroup>
  );
}
