import { useEffect, useState } from "react";
import { PageLayout } from "../PageLayout";
import { useSettingList, useSetSetting } from "../../api/settings";
import { GeneralSetting } from "./GeneralSetting";
import { AppearanceSetting } from "./AppearanceSetting";
import { DnsSetting } from "./DnsSetting";

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
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:gap-8">
        <GeneralSetting val={val} set={set} />
        <DnsSetting val={val} set={set} />
        <AppearanceSetting val={val} set={set} />
      </div>
    </PageLayout>
  );
}
