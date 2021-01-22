package tankax

import (
	"fmt"

	"github.com/grafana/tanka/pkg/tanka"

	"github.com/fatih/color"

	"github.com/grafana/tanka/pkg/kubernetes"
	"github.com/grafana/tanka/pkg/kubernetes/client"
	"github.com/grafana/tanka/pkg/term"
)

func Apply(l *tanka.LoadResult, opts tanka.ApplyOpts) error {
	kube, err := l.Connect()
	if err != nil {
		return err
	}
	defer kube.Close()

	// show diff
	diff, err := kube.Diff(l.Resources, kubernetes.DiffOpts{Strategy: opts.DiffStrategy})
	switch {
	case err != nil:
		// This is not fatal, the diff is not strictly required
		fmt.Println("Error diffing:", err)
	case diff == nil:
		tmp := "Warning: There are no differences. Your apply may not do anything at all."
		diff = &tmp
	}

	// in case of non-fatal error diff may be nil
	if diff != nil {
		b := term.Colordiff(*diff)
		fmt.Print(b.String())
	}

	// prompt for confirmation
	if opts.AutoApprove {
	} else if err := confirmPrompt("Applying to", l.Env.Spec.Namespace, kube.Info()); err != nil {
		return err
	}

	return kube.Apply(l.Resources, kubernetes.ApplyOpts{
		Force:    opts.Force,
		Validate: opts.Validate,
	})
}

// confirmPrompt asks the user for confirmation before apply
func confirmPrompt(action, namespace string, info client.Info) error {
	alert := color.New(color.FgRed, color.Bold).SprintFunc()

	return term.Confirm(
		fmt.Sprintf(`%s namespace '%s' of cluster '%s' at '%s' using context '%s'.`, action,
			alert(namespace),
			alert(info.Kubeconfig.Cluster.Name),
			alert(info.Kubeconfig.Cluster.Cluster.Server),
			alert(info.Kubeconfig.Context.Name),
		),
		"yes",
	)
}

func Delete(l *tanka.LoadResult, opts tanka.DeleteOpts) error {
	kube, err := l.Connect()
	if err != nil {
		return err
	}
	defer kube.Close()

	// show diff
	// static differ will never fail and always return something if input is not nil
	diff, err := kubernetes.StaticDiffer(false)(l.Resources)

	if err != nil {
		fmt.Println("Error diffing:", err)
	}

	// in case of non-fatal error diff may be nil
	if diff != nil {
		b := term.Colordiff(*diff)
		fmt.Print(b.String())
	}

	// prompt for confirmation
	if !opts.AutoApprove {
		if err := confirmPrompt("Deleting from", l.Env.Spec.Namespace, kube.Info()); err != nil {
			return err
		}
	}

	return kube.Delete(l.Resources, kubernetes.DeleteOpts{
		Force:    opts.Force,
		Validate: opts.Validate,
	})
}

func Prune(l *tanka.LoadResult, opts tanka.PruneOpts) error {
	kube, err := l.Connect()
	if err != nil {
		return err
	}
	defer kube.Close()

	// find orphaned resources
	orphaned, err := kube.Orphaned(l.Resources)
	if err != nil {
		return err
	}

	if len(orphaned) == 0 {
		fmt.Println("Nothing found to prune.")
		return nil
	}

	// print diff
	diff, err := kubernetes.StaticDiffer(false)(orphaned)
	if err != nil {
		// static diff can't fail normally, so unlike in apply, this is fatal
		// here
		return err
	}
	fmt.Print(term.Colordiff(*diff).String())

	// prompt for confirm
	if opts.AutoApprove {
	} else if err := confirmPrompt("Pruning from", l.Env.Spec.Namespace, kube.Info()); err != nil {
		return err
	}

	// delete resources
	return kube.Delete(orphaned, kubernetes.DeleteOpts{
		Force: opts.Force,
	})
}
