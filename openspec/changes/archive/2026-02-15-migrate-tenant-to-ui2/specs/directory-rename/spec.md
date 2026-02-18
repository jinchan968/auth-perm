# directory-rename 规范

## 新增需求

### 需求:ui2必须重命名为ui
系统必须将ui2目录重命名为ui。

#### 场景:重命名ui2为ui
- **当** 执行重命名操作
- **那么** ui2 目录被重命名为 ui

### 需求:ui必须标记为废弃
系统必须将原ui目录重命名为ui-deprecated。

#### 场景:重命名ui为ui-deprecated
- **当** 执行重命名操作
- **那么** ui 目录被重命名为 ui-deprecated
