"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { Project } from "@multica/core/types";
import { useWorkspaceId } from "@multica/core/hooks";
import { memberListOptions } from "@multica/core/workspace/queries";
import { useUpdateProjectAccess } from "@multica/core/projects/mutations";
import { Button } from "@multica/ui/components/ui/button";
import { Checkbox } from "@multica/ui/components/ui/checkbox";
import { toast } from "sonner";
import { useT } from "../../i18n";

export function ProjectAccessSettings({ project }: { project: Project }) {
  const { t } = useT("projects");
  const wsId = useWorkspaceId();
  const { data: members = [] } = useQuery(memberListOptions(wsId));
  const update = useUpdateProjectAccess();
  const [restricted, setRestricted] = useState(project.access_restricted === true);
  const [selected, setSelected] = useState(project.allowed_user_ids ?? []);
  const [filter, setFilter] = useState("");
  if (project.can_manage_access !== true) return null;

  return (
    <section className="space-y-3" aria-label={t(($) => $.access.title)}>
      <h2 className="text-caption font-medium">{t(($) => $.access.title)}</h2>
      <label className="flex items-center gap-2 text-caption">
        <Checkbox checked={restricted} onCheckedChange={(checked) => setRestricted(checked === true)} disabled={update.isPending} />
        {t(($) => $.access.restricted)}
      </label>
      {restricted && <>
        <p className="text-caption text-muted-foreground">{t(($) => $.access.creator_hint)}</p>
        <input className="w-full rounded-md border bg-background px-2 py-1 text-caption" aria-label={t(($) => $.access.search)} placeholder={t(($) => $.access.search)} value={filter} onChange={(event) => setFilter(event.target.value)} />
        <div className="max-h-56 space-y-2 overflow-y-auto">
          {members.filter((member) => member.name.toLowerCase().includes(filter.toLowerCase())).map((member) => {
            const creator = member.user_id === project.created_by;
            return <label key={member.user_id} className="flex items-center gap-2 text-caption">
              <Checkbox checked={creator || selected.includes(member.user_id)} disabled={creator || update.isPending} onCheckedChange={(checked) => setSelected((old) => checked ? [...old, member.user_id] : old.filter((id) => id !== member.user_id))} />
              <span className="truncate">{member.name}{creator ? ` (${t(($) => $.access.creator)})` : ""}</span>
            </label>;
          })}
        </div>
      </>}
      <Button size="sm" disabled={update.isPending} aria-busy={update.isPending} onClick={() => update.mutate({ id: project.id, access_restricted: restricted, allowed_user_ids: selected }, {
        onSuccess: () => toast.success(t(($) => $.access.saved)),
        onError: () => toast.error(t(($) => $.access.failed)),
      })}>{t(($) => $.access.save)}</Button>
    </section>
  );
}
