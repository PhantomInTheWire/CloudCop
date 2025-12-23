"use client";

import { PageHeader } from "@/components/page-header";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import {
  BellIcon,
  PaletteIcon,
  ShieldIcon,
  KeyIcon,
  EnvelopeIcon,
} from "@phosphor-icons/react";

export default function SettingsPage() {
  return (
    <>
      <PageHeader title="Settings" breadcrumbs={[{ label: "Settings" }]} />

      <div className="flex flex-1 flex-col gap-6 p-6 max-w-4xl">
        {/* General Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <PaletteIcon className="size-5" />
              Appearance
            </CardTitle>
            <CardDescription>
              Customize how CloudCop looks and feels
            </CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel>Theme</FieldLabel>
                <Select defaultValue="system">
                  <SelectTrigger className="w-[200px]">
                    <SelectValue placeholder="Select theme" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="light">Light</SelectItem>
                    <SelectItem value="dark">Dark</SelectItem>
                    <SelectItem value="system">System</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BellIcon className="size-5" />
              Notifications
            </CardTitle>
            <CardDescription>
              Configure how you receive security alerts
            </CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel>Email Notifications</FieldLabel>
                <Select defaultValue="critical">
                  <SelectTrigger className="w-[250px]">
                    <SelectValue placeholder="Select frequency" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All findings</SelectItem>
                    <SelectItem value="high">High and Critical only</SelectItem>
                    <SelectItem value="critical">Critical only</SelectItem>
                    <SelectItem value="none">Disabled</SelectItem>
                  </SelectContent>
                </Select>
              </Field>

              <Field>
                <FieldLabel>Notification Email</FieldLabel>
                <Input
                  type="email"
                  placeholder="security@example.com"
                  className="max-w-md"
                />
              </Field>

              <div className="pt-4">
                <h4 className="text-sm font-medium mb-3">Integrations</h4>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 rounded-lg border">
                    <div className="flex items-center gap-3">
                      <div className="size-10 rounded bg-muted flex items-center justify-center">
                        <EnvelopeIcon className="size-5" />
                      </div>
                      <div>
                        <div className="font-medium">Slack</div>
                        <div className="text-sm text-muted-foreground">
                          Send alerts to a Slack channel
                        </div>
                      </div>
                    </div>
                    <Badge variant="secondary">Coming Soon</Badge>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-lg border">
                    <div className="flex items-center gap-3">
                      <div className="size-10 rounded bg-muted flex items-center justify-center">
                        <BellIcon className="size-5" />
                      </div>
                      <div>
                        <div className="font-medium">PagerDuty</div>
                        <div className="text-sm text-muted-foreground">
                          Create incidents for critical findings
                        </div>
                      </div>
                    </div>
                    <Badge variant="secondary">Coming Soon</Badge>
                  </div>
                </div>
              </div>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Security Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShieldIcon className="size-5" />
              Security
            </CardTitle>
            <CardDescription>
              Manage your account security settings
            </CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <div className="flex items-center justify-between p-3 rounded-lg border">
                <div className="flex items-center gap-3">
                  <div className="size-10 rounded bg-muted flex items-center justify-center">
                    <KeyIcon className="size-5" />
                  </div>
                  <div>
                    <div className="font-medium">Two-Factor Authentication</div>
                    <div className="text-sm text-muted-foreground">
                      Add an extra layer of security to your account
                    </div>
                  </div>
                </div>
                <Button variant="outline">Enable</Button>
              </div>

              <div className="flex items-center justify-between p-3 rounded-lg border">
                <div>
                  <div className="font-medium">API Keys</div>
                  <div className="text-sm text-muted-foreground">
                    Manage API keys for programmatic access
                  </div>
                </div>
                <Button variant="outline">Manage Keys</Button>
              </div>

              <div className="flex items-center justify-between p-3 rounded-lg border">
                <div>
                  <div className="font-medium">Session Timeout</div>
                  <div className="text-sm text-muted-foreground">
                    Automatically log out after inactivity
                  </div>
                </div>
                <Select defaultValue="60">
                  <SelectTrigger className="w-[150px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="30">30 minutes</SelectItem>
                    <SelectItem value="60">1 hour</SelectItem>
                    <SelectItem value="240">4 hours</SelectItem>
                    <SelectItem value="480">8 hours</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </FieldGroup>
          </CardContent>
        </Card>

        {/* Save Button */}
        <div className="flex justify-end">
          <Button>Save Changes</Button>
        </div>
      </div>
    </>
  );
}
