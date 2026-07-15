import { useEffect, useState } from "react";
import { PageLayout } from "../PageLayout";
import { useSettingList, useSetSetting } from "../../api/settings";
import { General } from "./General";
import { Appearance } from "./Appearance";
import { Dns } from "./Dns";
import { Storage } from "./Storage";
import { Tag } from "./Tag";
import { Rename } from "./Rename";

export default function Setting() {
  const { data, isLoading } = useSettingList();
  const setSetting = useSetSetting();
  const [local, setLocal] = useState<Record<string, string>>({});

  useEffect(() => {
    if (data) {
      const map: Record<string, string> = {};
      for (const s of data) map[s.key] = s.value;
      setLocal(map);
    }
  }, [data]);

  const set = (key: string, value: string) => {
    setLocal((prev) => ({ ...prev, [key]: value }));
    setSetting.mutate({ key, value });
  };

  const val = (key: string) => local[key] ?? "";

  return (
    <PageLayout title="设置">
      <div className="columns-1 gap-4 lg:columns-2 lg:gap-8 [&>.settings-category]:mb-4 [&>.settings-category]:break-inside-avoid lg:[&>.settings-category]:mb-8">
        <Appearance val={val} set={set} />
        <General val={val} set={set} />
        <Dns val={val} set={set} />
        <Storage />
        <Rename />
        <Tag />
      </div>
    </PageLayout>
  );
}
