# AuthZEN 情報フロー図

このドキュメントでは、AuthZEN（Authorization API）における情報フローとデータモデルを図で説明します。

## 1. 認可リクエスト・レスポンスの構造

AuthZENの認可リクエストとレスポンスの基本構造を示します。

```mermaid
graph TD
    subgraph "認可リクエスト"
        Subject["Subject\n- type: string\n- id: string\n- properties: object"]
        Resource["Resource\n- type: string\n- id: string\n- properties: object"]
        Action["Action\n- name: string\n- properties: object"]
        Context["Context\n- 任意のJSON属性"]
        
        Request["AuthorizeRequest"]
        Request --> Subject
        Request --> Resource
        Request --> Action
        Request --> Context
    end
    
    subgraph "認可レスポンス"
        Decision["Decision\n- decision: boolean"]
        ResponseContext["Context\n- 任意のJSON属性"]
        
        Response["AuthorizeResponse"]
        Response --> Decision
        Response --> ResponseContext
    end
    
    Request -.-> Response
```

## 2. 複数認可リクエスト（Evaluations）の構造

複数の認可判断を一度に行うためのリクエスト構造を示します。

```mermaid
graph TD
    subgraph "複数認可リクエスト"
        DefaultSubject["Default Subject"]
        DefaultAction["Default Action"]
        DefaultContext["Default Context"]
        
        Eval1["Evaluation 1\n- Resource 1"]
        Eval2["Evaluation 2\n- Resource 2\n- Action 2"]
        Eval3["Evaluation 3\n- Resource 3"]
        
        Options["Options\n- evaluations_semantic: string"]
        
        EvaluationsRequest["EvaluationsRequest"]
        EvaluationsRequest --> DefaultSubject
        EvaluationsRequest --> DefaultAction
        EvaluationsRequest --> DefaultContext
        EvaluationsRequest --> Eval1
        EvaluationsRequest --> Eval2
        EvaluationsRequest --> Eval3
        EvaluationsRequest --> Options
    end
    
    subgraph "複数認可レスポンス"
        EvalResult1["Result 1\n- decision: boolean"]
        EvalResult2["Result 2\n- decision: boolean"]
        EvalResult3["Result 3\n- decision: boolean"]
        
        EvaluationsResponse["EvaluationsResponse"]
        EvaluationsResponse --> EvalResult1
        EvaluationsResponse --> EvalResult2
        EvaluationsResponse --> EvalResult3
    end
    
    EvaluationsRequest -.-> EvaluationsResponse
```

## 3. 検索APIのデータフロー

### 3.1 Subject Search

```mermaid
graph LR
    subgraph "Subject Search リクエスト"
        SS_Subject["Subject\n- type: string"]
        SS_Resource["Resource\n- type: string\n- id: string"]
        SS_Action["Action\n- name: string"]
        
        SS_Request["SubjectSearchRequest"]
        SS_Request --> SS_Subject
        SS_Request --> SS_Resource
        SS_Request --> SS_Action
    end
    
    subgraph "Subject Search レスポンス"
        SS_Results["Results\n- Subject[]"]
        SS_Page["Page\n- next_token: string"]
        
        SS_Response["SubjectSearchResponse"]
        SS_Response --> SS_Results
        SS_Response --> SS_Page
    end
    
    SS_Request -.-> SS_Response
```

### 3.2 Resource Search

```mermaid
graph LR
    subgraph "Resource Search リクエスト"
        RS_Subject["Subject\n- type: string\n- id: string"]
        RS_Resource["Resource\n- type: string"]
        RS_Action["Action\n- name: string"]
        
        RS_Request["ResourceSearchRequest"]
        RS_Request --> RS_Subject
        RS_Request --> RS_Resource
        RS_Request --> RS_Action
    end
    
    subgraph "Resource Search レスポンス"
        RS_Results["Results\n- Resource[]"]
        RS_Page["Page\n- next_token: string"]
        
        RS_Response["ResourceSearchResponse"]
        RS_Response --> RS_Results
        RS_Response --> RS_Page
    end
    
    RS_Request -.-> RS_Response
```

### 3.3 Action Search

```mermaid
graph LR
    subgraph "Action Search リクエスト"
        AS_Subject["Subject\n- type: string\n- id: string"]
        AS_Resource["Resource\n- type: string\n- id: string"]
        
        AS_Request["ActionSearchRequest"]
        AS_Request --> AS_Subject
        AS_Request --> AS_Resource
    end
    
    subgraph "Action Search レスポンス"
        AS_Results["Results\n- Action[]"]
        AS_Page["Page\n- next_token: string"]
        
        AS_Response["ActionSearchResponse"]
        AS_Response --> AS_Results
        AS_Response --> AS_Page
    end
    
    AS_Request -.-> AS_Response
```

## 4. メタデータディスカバリーの構造

```mermaid
graph TD
    subgraph "メタデータレスポンス"
        PDP["policy_decision_point\n- string"]
        AccessEval["access_evaluation_endpoint\n- string"]
        AccessEvals["access_evaluations_endpoint\n- string (optional)"]
        SearchSubject["search_subject_endpoint\n- string (optional)"]
        SearchResource["search_resource_endpoint\n- string (optional)"]
        SearchAction["search_action_endpoint\n- string (optional)"]
        
        MetadataResponse["MetadataResponse"]
        MetadataResponse --> PDP
        MetadataResponse --> AccessEval
        MetadataResponse --> AccessEvals
        MetadataResponse --> SearchSubject
        MetadataResponse --> SearchResource
        MetadataResponse --> SearchAction
    end
```

## 5. エラーレスポンスの構造

```mermaid
graph TD
    subgraph "エラーレスポンス"
        Status["HTTP Status Code\n- 400: Bad Request\n- 401: Unauthorized\n- 403: Forbidden\n- 500: Internal Error"]
        
        ErrorBody["Error Body\n- エラーメッセージ"]
        
        ErrorResponse["Error Response"]
        ErrorResponse --> Status
        ErrorResponse --> ErrorBody
    end
```

## 6. 認可判断のロジックフロー

```mermaid
flowchart TD
    Start([開始]) --> InputData[/Subject, Resource, Action, Context/]
    InputData --> ValidateInput{入力検証}
    ValidateInput -->|無効| ReturnError[エラーレスポンス]
    ValidateInput -->|有効| FindPolicy[ポリシー検索]
    FindPolicy --> PolicyFound{ポリシー発見?}
    PolicyFound -->|Yes| EvaluatePolicy[ポリシー評価]
    PolicyFound -->|No| DefaultDeny[デフォルト拒否]
    EvaluatePolicy --> Decision{決定}
    Decision -->|許可| Allow[許可レスポンス]
    Decision -->|拒否| Deny[拒否レスポンス]
    DefaultDeny --> Deny
    Allow --> End([終了])
    Deny --> End
    ReturnError --> End
```

## 7. HTTPSバインディングの例

```mermaid
sequenceDiagram
    participant PEP as Policy Enforcement Point
    participant PDP as Policy Decision Point
    
    PEP->>+PDP: POST /access/v1/evaluation HTTP/1.1<br>Host: pdp.example.com<br>Authorization: Bearer token123<br>Content-Type: application/json<br><br>{subject, resource, action, context}
    
    PDP-->>-PEP: HTTP/1.1 200 OK<br>Content-Type: application/json<br><br>{decision: true}
