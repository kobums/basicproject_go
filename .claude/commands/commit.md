---
description: '현재까지의 작업을 논리적 단위별로 나눠서 커밋'
---

현재까지 작업한 내용을 작업 단위별로 나눠서 커밋해주세요.

## 수행 단계

1. **현재 변경사항 확인**
   - `git status`로 staged/unstaged 변경사항 확인

   - `git diff`로 각 파일의 변경 내용 파악

2. **작업을 논리적 단위로 분류**
   - 관련된 변경사항끼리 그룹화

   - 각 그룹이 하나의 완성된 기능/수정 단위가 되도록 구성

   - 예: 기능 추가, 버그 수정, 리팩토링, 문서 수정 등으로 분류

3. **각 단위별로 순차 커밋**
   - 가장 핵심적인 변경부터 시작

   - 각 작업 단위별로:
     - `git add [관련 파일들]`

     - 컨벤션에 맞는 커밋 메시지 작성

     - `git commit -m "type(scope): 설명"`

     ## 커밋 메시지 형식
     - 형식: `git commit -m "type(scope): 간단한 설명"`
     - Co-Authored-By 등의 작성자 정보는 추가하지 않음
     - heredoc(<<EOF) 사용하지 않고 -m 옵션으로 직접 작성
     - 예시: `git commit -m "feat(auth): 소셜 로그인 기능 추가"`

   - 커밋 타입: feat, fix, refactor, docs, style, test, chore

4. **커밋 완료 확인**
   - `git log --oneline -10`으로 커밋 이력 확인

   - `git status`로 남은 변경사항 확인

## 주의사항

- 각 커밋은 독립적으로 의미가 있어야 함

- 관련 없는 변경사항은 별도 커밋으로 분리

- .env, credentials 등 민감한 파일은 커밋하지 않음

- **커밋 메시지에 Co-Authored-By 등 작성자 정보를 추가하지 않음**

- **$(cat <<'EOF' ... EOF) 같은 heredoc 문법 사용하지 않고 직접 -m 옵션으로 메시지 작성**
