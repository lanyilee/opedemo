package controllers

import (
	cachev1 "github.com/lanyilee/opedemo/apis/cache/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// 根据自定义的appservice 创建相应的deployment,分为typemeta，objectmeta,spec三部分
// typemeta,objectmeta都是用meta包，spec用到appsv1包和corev1包（pod部分）
func NewDeploy(app *cachev1.AppService) *appsv1.Deployment {
	labels := map[string]string{"app": app.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, schema.GroupVersionKind{
					Group:   appsv1.SchemeGroupVersion.Group,
					Version: appsv1.SchemeGroupVersion.Version,
					Kind:    "AppService",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: app.Spec.Size,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: NewContainer(app),
				},
			},
		},
	}

}

// 新建容器，基本都是corev1包，注意corev1.ContainerPort{} 包含了ContainerPort这个属性
func NewContainer(app *cachev1.AppService) []corev1.Container {
	ports := []corev1.ContainerPort{}
	for _, p := range app.Spec.Ports {
		c := corev1.ContainerPort{}
		c.ContainerPort = p.TargetPort.IntVal
		ports = append(ports, c)
	}
	return []corev1.Container{
		{
			Name:            app.Name,
			Image:           app.Spec.Image,
			Resources:       app.Spec.Resources,
			Env:             app.Spec.Envs,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports:           ports,
		},
	}
}

// 新建服务service,也是分为typemeta,objectmeta,spec三部分,注意spec的selector的labels要和上面deployment设置的一样才能匹配
func NewService(app *cachev1.AppService) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",      // 版本号
			Kind:       "Service", // 大写
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&app, schema.GroupVersionKind{
					Group:   appsv1.SchemeGroupVersion.Group,
					Version: appsv1.SchemeGroupVersion.Version,
					Kind:    "AppService",
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Ports:    app.Spec.Ports,
			Selector: map[string]string{"app": app.Name},
		},
	}
}
