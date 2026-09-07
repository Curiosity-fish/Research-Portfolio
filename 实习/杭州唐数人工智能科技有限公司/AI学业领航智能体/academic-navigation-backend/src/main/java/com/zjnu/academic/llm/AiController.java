package com.zjnu.academic.llm;

import com.zjnu.academic.common.result.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@Tag(name = "AI 解读", description = "调用本地大模型对指标数据进行解读，模型不可用时自动降级为规则摘要")
@RestController
@RequestMapping("/api/ai")
public class AiController {

    private final AiInterpretService aiInterpretService;

    @Autowired
    public AiController(AiInterpretService aiInterpretService) {
        this.aiInterpretService = aiInterpretService;
    }

    @Operation(summary = "指标解读", description = "传入指标 JSON 与可选的解读指令，返回 AI 解读文本")
    @PostMapping("/interpret")
    public Result<Map<String, Object>> interpret(@RequestBody AiInterpretRequest request) {
        return Result.ok(aiInterpretService.interpret(request.instruction(), request.data()));
    }
}
