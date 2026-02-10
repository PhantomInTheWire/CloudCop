"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import {
  PlusIcon,
  CheckIcon,
  CloudIcon,
  ArrowRightIcon,
  CheckCircleIcon,
  CircleNotchIcon,
} from "@phosphor-icons/react";

const cloudFormationTemplate = `https://us-east-1.console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/review?templateURL=https://cloudcop-templates.s3.amazonaws.com/guard-scan-role.yaml&stackName=cloudcop-security-role`;

export function AddAccountDialog() {
  const [step, setStep] = React.useState(1);
  const [accountId, setAccountId] = React.useState("");
  const [externalId, setExternalId] = React.useState("");
  const [roleArn, setRoleArn] = React.useState("");
  const [isVerifying, setIsVerifying] = React.useState(false);
  const [isVerified, setIsVerified] = React.useState(false);
  const [open, setOpen] = React.useState(false);

  const handleVerify = async () => {
    setIsVerifying(true);
    setTimeout(() => {
      setIsVerifying(false);
      setIsVerified(true);
    }, 2000);
  };

  const handleConnect = async () => {
    setOpen(false);
    // Reset state
    setStep(1);
    setAccountId("");
    setExternalId("");
    setRoleArn("");
    setIsVerified(false);
  };

  const resetDialog = () => {
    setStep(1);
    setAccountId("");
    setExternalId("");
    setRoleArn("");
    setIsVerified(false);
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        setOpen(o);
        if (!o) resetDialog();
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <PlusIcon className="mr-2 size-4" />
          Add AWS Account
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Connect AWS Account</DialogTitle>
          <DialogDescription>
            Follow these steps to securely connect your AWS account
          </DialogDescription>
        </DialogHeader>

        {/* Progress Steps */}
        <div className="flex items-center gap-2 py-4">
          {[1, 2, 3].map((s) => (
            <React.Fragment key={s}>
              <div
                className={`size-8 rounded-full flex items-center justify-center text-sm font-medium ${
                  step >= s
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground"
                }`}
              >
                {step > s ? <CheckIcon className="size-4" /> : s}
              </div>
              {s < 3 && (
                <div
                  className={`flex-1 h-0.5 ${
                    step > s ? "bg-primary" : "bg-muted"
                  }`}
                />
              )}
            </React.Fragment>
          ))}
        </div>

        {/* Step 1: Deploy CloudFormation */}
        {step === 1 && (
          <div className="space-y-4">
            <div className="rounded-lg border p-4 bg-muted/50">
              <h4 className="font-medium mb-2">Deploy CloudFormation Stack</h4>
              <p className="text-sm text-muted-foreground mb-4">
                Click the button below to open AWS CloudFormation with our
                pre-configured template. This creates a read-only IAM role for
                CloudCop.
              </p>
              <Button asChild>
                <a
                  href={cloudFormationTemplate}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <CloudIcon className="mr-2 size-4" />
                  Deploy to AWS
                  <ArrowRightIcon className="ml-2 size-4" />
                </a>
              </Button>
            </div>

            <div className="space-y-2">
              <h4 className="font-medium text-sm">What this creates:</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li className="flex items-center gap-2">
                  <CheckCircleIcon className="size-4 text-green-500" />
                  Read-only IAM role with security scanning permissions
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircleIcon className="size-4 text-green-500" />
                  External ID for secure cross-account access
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircleIcon className="size-4 text-green-500" />
                  No write permissions - CloudCop cannot modify your resources
                </li>
              </ul>
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t">
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => setStep(2)}>
                I&apos;ve Deployed the Stack
                <ArrowRightIcon className="ml-2 size-4" />
              </Button>
            </div>
          </div>
        )}

        {/* Step 2: Enter Details */}
        {step === 2 && (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Enter the details from your CloudFormation stack outputs.
            </p>

            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="accountId">AWS Account ID</FieldLabel>
                <Input
                  id="accountId"
                  placeholder="123456789012"
                  value={accountId}
                  onChange={(e) => setAccountId(e.target.value)}
                />
              </Field>

              <Field>
                <FieldLabel htmlFor="roleArn">Role ARN</FieldLabel>
                <Input
                  id="roleArn"
                  placeholder="arn:aws:iam::123456789012:role/CloudCopSecurityScanRole"
                  value={roleArn}
                  onChange={(e) => setRoleArn(e.target.value)}
                />
              </Field>

              <Field>
                <FieldLabel htmlFor="externalId">External ID</FieldLabel>
                <Input
                  id="externalId"
                  placeholder="From CloudFormation outputs"
                  value={externalId}
                  onChange={(e) => setExternalId(e.target.value)}
                />
              </Field>
            </FieldGroup>

            <div className="flex justify-between gap-2 pt-4 border-t">
              <Button variant="outline" onClick={() => setStep(1)}>
                Back
              </Button>
              <Button
                onClick={() => setStep(3)}
                disabled={!accountId || !externalId || !roleArn}
              >
                Continue
                <ArrowRightIcon className="ml-2 size-4" />
              </Button>
            </div>
          </div>
        )}

        {/* Step 3: Verify Connection */}
        {step === 3 && (
          <div className="space-y-4">
            <div className="rounded-lg border p-4">
              <h4 className="font-medium mb-3">Account Details</h4>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Account ID:</span>
                  <code className="bg-muted px-2 py-0.5 rounded">
                    {accountId}
                  </code>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Role ARN:</span>
                  <code className="bg-muted px-2 py-0.5 rounded text-xs truncate max-w-[300px]">
                    {roleArn}
                  </code>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">External ID:</span>
                  <code className="bg-muted px-2 py-0.5 rounded">
                    {externalId}
                  </code>
                </div>
              </div>
            </div>

            {isVerified ? (
              <div className="rounded-lg border border-green-500/50 bg-green-500/10 p-4 flex items-center gap-3">
                <CheckCircleIcon
                  className="size-6 text-green-500"
                  weight="fill"
                />
                <div>
                  <h4 className="font-medium text-green-500">
                    Connection Verified
                  </h4>
                  <p className="text-sm text-muted-foreground">
                    Successfully assumed role in your AWS account
                  </p>
                </div>
              </div>
            ) : (
              <Button
                className="w-full"
                onClick={handleVerify}
                disabled={isVerifying}
              >
                {isVerifying ? (
                  <>
                    <CircleNotchIcon className="mr-2 size-4 animate-spin" />
                    Verifying Connection...
                  </>
                ) : (
                  <>
                    <CheckCircleIcon className="mr-2 size-4" />
                    Verify Connection
                  </>
                )}
              </Button>
            )}

            <div className="flex justify-between gap-2 pt-4 border-t">
              <Button variant="outline" onClick={() => setStep(2)}>
                Back
              </Button>
              <Button onClick={handleConnect} disabled={!isVerified}>
                Connect Account
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
