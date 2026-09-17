import { beforeEach, describe, expect, it, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Project } from "@multica/core/types";
import { renderWithI18n } from "../../test/i18n";
import { ProjectAccessSettings } from "./project-access-settings";

const mocks = vi.hoisted(() => ({ mutate: vi.fn(), pending: false }));
vi.mock("@tanstack/react-query", () => ({ useQuery: () => ({ data: [{ user_id: "creator", name: "Creator" }, { user_id: "member", name: "Member" }] }) }));
vi.mock("@multica/core/hooks", () => ({ useWorkspaceId: () => "workspace" }));
vi.mock("@multica/core/workspace/queries", () => ({ memberListOptions: () => ({}) }));
vi.mock("@multica/core/projects/mutations", () => ({ useUpdateProjectAccess: () => ({ mutate: mocks.mutate, isPending: mocks.pending }) }));
const project: Project = { id: "project", workspace_id: "workspace", title: "Project", description: null, icon: null, status: "planned", priority: "none", lead_type: null, lead_id: null, start_date: null, due_date: null, created_at: "", updated_at: "", issue_count: 0, done_count: 0, resource_count: 0, created_by: "creator", can_manage_access: true };
beforeEach(() => { mocks.mutate.mockReset(); mocks.pending = false; });
describe("ProjectAccessSettings", () => {
  it("saves the selected member list only after Save is pressed", async () => {
    const user = userEvent.setup();
    renderWithI18n(<ProjectAccessSettings project={project} />);
    await user.click(screen.getByRole("checkbox", { name: "Only selected members" }));
    expect(screen.getByRole("checkbox", { name: "Creator (Creator)" })).toHaveAttribute("aria-disabled", "true");
    await user.click(screen.getByRole("checkbox", { name: "Member" }));
    expect(mocks.mutate).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Save access" }));
    expect(mocks.mutate).toHaveBeenCalledWith({ id: "project", access_restricted: true, allowed_user_ids: ["member"] }, expect.any(Object));
  });
  it("does not offer access management to a non-creator", () => {
    renderWithI18n(<ProjectAccessSettings project={{ ...project, can_manage_access: false }} />);
    expect(screen.queryByRole("button", { name: "Save access" })).not.toBeInTheDocument();
  });
  it("disables Save while the request is pending", () => {
    mocks.pending = true;
    renderWithI18n(<ProjectAccessSettings project={project} />);
    expect(screen.getByRole("button", { name: "Save access" })).toBeDisabled();
  });
});
