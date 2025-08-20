package main

import (
	"fmt"
	"log"
	"math"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/net/context"
	"google.golang.org/genai"
)

type GeminiEmbedder struct {
	Client *genai.Client
}

func NewGeminiEmbedder() embedding.Embedder {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &GeminiEmbedder{
		Client: client,
	}
}

// cosineSimilarity calculates the similarity between two vectors.
func cosineSimilarity(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have the same length")
	}

	var dotProduct, aMagnitude, bMagnitude float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		aMagnitude += float64(a[i] * a[i])
		bMagnitude += float64(b[i] * b[i])
	}

	if aMagnitude == 0 || bMagnitude == 0 {
		return 0, nil
	}

	return dotProduct / (math.Sqrt(aMagnitude) * math.Sqrt(bMagnitude)), nil
}

// EmbedStrings 将多条文本转换成向量
func (e *GeminiEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	var contents []*genai.Content
	for _, text := range texts {
		contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
	}

	result, err := e.Client.Models.EmbedContent(ctx,
		"gemini-embedding-001",
		contents,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("embedding error: %w", err)
	}

	// 转换成 [][]float64
	embeddings := make([][]float64, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		vec := make([]float64, len(emb.Values))
		for j, v := range emb.Values {
			vec[j] = float64(v)
		}
		embeddings[i] = vec
	}

	// 如果你还想计算相似度，可以放在这里
	for i := 0; i < len(texts); i++ {
		for j := i + 1; j < len(texts); j++ {
			similarity, _ := cosineSimilarity(
				result.Embeddings[i].Values,
				result.Embeddings[j].Values,
			)
			fmt.Printf("Similarity between '%s' and '%s': %.4f\n",
				texts[i], texts[j], similarity)
		}
	}

	return embeddings, nil
}

func main() {
	ctx := context.Background()

	// 初始化嵌入器（示例使用）
	embedder := NewGeminiEmbedder()

	//embedder.EmbedStrings()
	// 初始化分割器
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,
		BufferSize:   2,
		MinChunkSize: 100,
		Separators:   []string{"\n", "。", ".","?", "!"},
		Percentile:   0.9,
	})
	if err != nil {
		panic(err)
	}

	// 准备要分割的文档
	docs := []*schema.Document{
		{
			ID:      "doc1",
			Content: `雅思阅读考点词 库—引用于刘洪波《剑桥雅思阅读考点词真经》538 个雅思阅读考点词，分成了 3 类，进行了重要性排序。如果考生熟知，对应雅思阅读 7 分以上能力。第 1 类考点词共 20+34=54 个定义：雅思阅读文章中只要出现该次词，90％会被命题考查！不同的真题中被反复考查多次，超高频考点词！背记：考生请严格按下表中重要性排行顺序记忆 20 个考点词。第一个背的阅读单词是 resemble 而不是 abandon,它在《剑桥雅思 9》和《剑桥雅思 10》中多次考到。请同记忆下表真题命题方式中相对于的 34 个同义替换考点词（彩色单词）。不熟的单词请一定查出标注，同时建议思考相关反义词表达。要求：滚瓜烂熟重要性排行考点词常考中文词义雅思阅读真题命题方式1resemblev.像，与……相似like,look,like,be similar to2recognizev.认出，识别；承认perceive,acknowledge,realize,appreciate, admit ,identify, comprehend, understandknow3adjustv.调整，使适合change, modify, shift, alter4approachn.方法method, way5fundamentaladj.基本的，基础的rudimentary, preliminary, basic6rely on依靠，依赖depend on7domesticadj.家庭的；国内的home, local, national8measurev.测量calculate, assess, evaluate9traitn.特性，特征characteristic, feature, property10coinv.创造first used, invent11artificialadj.人造的，仿造的synthetic, man-made12promptv.促进，激起initiate, immediately13exchangev.交换share, apply A to B14underliev.成为……基础based on, ground, root15ignorev.忽视，不顾neglect, overlook, underestimate\n重要性排行考点词常考中文词义雅思阅读真题命题方式16fertilisern.化肥，肥料chemical, toxic, unnatural17that*pron.那；那个this, it, they, those, these, such*指代是雅思阅读的重要考点18and*conj.和，而且or, as well as, both…and, not only…but also…, other than, in addition, besides, on the one hand…on the other hand…, neither…nor…*并列结构是雅思阅读的重要考点19rather than*而非，不是but, yet, however, whereas, nonetheless,nevertheless, although, notwithstandingthough, instead*转折结构是雅思阅读的重要考点20thanks to*由于，幸亏stem from，derive，owing to，due to,according to, because of, on account of,as a result of, leading to, because, since,for, in that, as, therefore, hence*因果关系是雅思阅读重要考点第二类考点词共 100+71=171 个（注：这类词中考点词与近义词有重复，故实际为 171 个考点词。）定义：雅思阅读文章中只要出现该词，60％会被命题考查！不同的真题中被考查过一次以上重要考点词！背记：考生请按下表中重要性排行顺序记忆 100 个考点词; 同时记忆真题命题方式中相对于的 71 个同意替换考点词（彩色单词）。不熟的单词请一定查出标注，同时建议思考相关反义词表达。要求：熟记 10 遍以上。重要性排行考点词常考中文词义雅思阅读真题命题方式21diversityn.多样性，差异variety, difference22detectv.查明，发现find, look for, seek, search23isolatev.使隔离，使孤立inaccessible24avoidv.避免escape, evitable25budgetn.预算fund, financial26adapt to使适应fit, suit27alternativeadj.供替代的，供选择的n.替代品substitute28compensaten.补偿，赔偿make up, offset\n重要性排行考点词常考中文词义雅思阅读真题命题方式29componentn.成分，要素proportion30militaryadj.军事的weapon, army31criterian.标准standard32curriculumn.课程syllabus, course of study33feasibleadj.可行的realistic, viable34constrainv.束缚，限制stop, control35deficiencyn.缺陷，缺点shortage, defect, weakness36supplementv.补充provision37distinguishv.区别，辨别separate, differentiate38analyzev.分析，解释examine, diagnose39emphasizev.强调，着重focus on, stress40enormousadj.庞大的，巨大的massive, large41imitatev.模仿mimic, copy42impairv.削弱，`,
		},
	}

	// 执行分割
	results, err := splitter.Transform(ctx, docs)
	if err != nil {
		panic(err)
	}

	// log.Println(results)

	// 处理分割结果
	for i, doc := range results {
		println("片段", i+1, ":", doc.Content)
	}
}
