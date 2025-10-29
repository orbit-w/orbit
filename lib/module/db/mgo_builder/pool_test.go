package mgo_builder

import (
	"sync"
	"testing"
)

// TestMongoUpdateBuilderPool 测试 MongoUpdateBuilder 对象池
func TestMongoUpdateBuilderPool(t *testing.T) {
	pool := NewMongoUpdateBuilderPool()

	// 测试基本的 Get/Put 操作
	builder := pool.Get()
	if builder == nil {
		t.Fatal("池应该返回非空的 builder")
	}

	// 使用 builder
	builder.Set("test", "value")
	if builder.IsEmpty() {
		t.Error("builder 应该不为空")
	}

	// 放回池中
	pool.Put(builder)

	// 再次获取，应该得到已重置的对象
	builder2 := pool.Get()
	if builder2 == nil {
		t.Fatal("池应该返回非空的 builder")
	}

	// 验证对象已被重置
	if !builder2.IsEmpty() {
		t.Error("从池中获取的对象应该是干净的")
	}

	pool.Put(builder2)
}

// TestGlobalPoolFunctions 测试全局池函数
func TestGlobalPoolFunctions(t *testing.T) {
	// 测试 GetBuilder/PutBuilder
	builder := GetBuilder()
	if builder == nil {
		t.Fatal("GetBuilder 应该返回非空的 builder")
	}

	builder.Set("global_test", "value")
	PutBuilder(builder)
}

// TestWithBuilder 测试 WithBuilder 函数
func TestWithBuilder(t *testing.T) {
	var result map[string]any

	WithBuilder(func(builder *MongoUpdateBuilder) {
		builder.Set("with_test", "value")
		result = builder.Build()
	})

	if result == nil {
		t.Error("WithBuilder 应该生成结果")
	}

	// 验证对象已被自动放回池中（通过再次获取验证重置）
	builder := GetBuilder()
	defer PutBuilder(builder)

	if !builder.IsEmpty() {
		t.Error("池中的对象应该是干净的")
	}
}

// TestWithBuilderResult 测试 WithBuilderResult 函数
func TestWithBuilderResult(t *testing.T) {
	result := WithBuilderResult(func(builder *MongoUpdateBuilder) map[string]any {
		builder.Set("result_test", "value")
		return builder.Build()
	})

	if result == nil {
		t.Error("WithBuilderResult 应该返回结果")
	}

	if result["$set"] == nil {
		t.Error("结果应该包含 $set 操作")
	}
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	const goroutines = 100
	const iterations = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 并发测试 Builder 池
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				WithBuilder(func(builder *MongoUpdateBuilder) {
					builder.Set("concurrent", j)
					_ = builder.Build()
				})
			}
		}()
	}

	wg.Wait()
}

// TestReset 测试 Reset 方法
func TestReset(t *testing.T) {
	// 测试 MongoUpdateBuilder.Reset()
	builder := NewMongoUpdateBuilder()
	builder.Set("test", "value")
	builder.Inc("counter", 1)

	if builder.IsEmpty() {
		t.Error("builder 应该不为空")
	}

	builder.Reset()

	if !builder.IsEmpty() {
		t.Error("Reset 后 builder 应该为空")
	}
}

// TestNilHandling 测试空指针处理
func TestNilHandling(t *testing.T) {
	pool := NewMongoUpdateBuilderPool()

	// Put nil 应该不会崩溃
	pool.Put(nil)
}

// BenchmarkPoolVsNew 基准测试：池 vs 直接创建
func BenchmarkPoolVsNew(b *testing.B) {
	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			builder := NewMongoUpdateBuilder()
			builder.Set("field", i)
			_ = builder.Build()
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			WithBuilder(func(builder *MongoUpdateBuilder) {
				builder.Set("field", i)
				_ = builder.Build()
			})
		}
	})
}
